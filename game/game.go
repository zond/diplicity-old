package game

import (
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/godip/classical"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/state"
	"github.com/zond/kcwraps/kol"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Minutes int

type Game struct {
	Id kol.Id

	Closed             bool `kol:"index"`
	State              common.GameState
	Variant            string
	AllocationMethod   string
	SecretEmail        common.SecretFlag
	SecretNickname     common.SecretFlag
	SecretNation       common.SecretFlag
	EndYear            int
	MinimumRanking     float64
	MaximumRanking     float64
	MinimumReliability float64
	Private            bool `kol:"index"`

	Deadlines map[dip.PhaseType]Minutes

	ChatFlags map[dip.PhaseType]common.ChatFlag

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (self *Game) Disallows(u *user.User) bool {
	return u.Ranking < self.MinimumRanking || u.Ranking > self.MaximumRanking || u.Reliability() < self.MinimumReliability
}

func (self *Game) allocate(d *kol.DB) error {
	members := self.Members(d)
	switch self.AllocationMethod {
	case common.RandomString:
		for memberIndex, nationIndex := range rand.Perm(len(members)) {
			members[memberIndex].Nation = common.VariantMap[self.Variant].Nations[nationIndex]
			if err := d.Set(&members[memberIndex]); err != nil {
				return err
			}
		}
		return nil
	case common.PreferencesString:
		prefs := make([][]dip.Nation, len(members))
		for index, member := range members {
			prefs[index] = member.PreferredNations
		}
		for index, nation := range optimizePreferences(prefs) {
			members[index].Nation = nation
			if err := d.Set(&members[index]); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Unknown allocation method %v", self.AllocationMethod)
}

func (self *Game) start(d *kol.DB) error {
	if self.State != common.GameStateCreated {
		return fmt.Errorf("%+v is already started", self)
	}
	self.State = common.GameStateStarted
	self.Closed = true
	if err := d.Set(self); err != nil {
		return err
	}
	if err := self.allocate(d); err != nil {
		return err
	}
	var startState *state.State
	if self.Variant == common.ClassicalString {
		startState = classical.Start()
	} else {
		return fmt.Errorf("Unknown variant %v", self.Variant)
	}
	startPhase := startState.Phase()
	phase := &Phase{
		GameId:  self.Id,
		Ordinal: 0,
		Season:  startPhase.Season(),
		Year:    startPhase.Year(),
		Type:    startPhase.Type(),
	}
	phase.Units, phase.SupplyCenters, phase.Dislodgeds, phase.Dislodgers, phase.Bounces = startState.Dump()
	return d.Set(phase)
}

func (self *Game) Updated(d *kol.DB, old *Game) {
	for _, member := range self.Members(d) {
		d.EmitUpdate(&member)
	}
}

func (self *Game) SanitizeMessage(sender *Member, message *Message) (result Message) {
	// the result is an anonymized message
	result = Message{
		Nation:    message.Nation,
		Body:      message.Body,
		Channel:   message.Channel,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
	}
	// if it is from anon, and it is a private channel, replace the sender with anon in the chanel
	if result.Nation == common.Anonymous && len(result.Channel) == 2 {
		delete(result.Channel, sender.Nation)
		result.Channel[common.Anonymous] = true
	}
	return
}

func (self *Game) MessageAllowed(phase *Phase, sender, recipient *Member, message *Message) bool {
	// disallow messages from those not in the channel
	if !message.Channel[sender.Nation] {
		return false
	}
	// disallow messages to those not in the channel
	if recipient != nil && !message.Channel[recipient.Nation] {
		return false
	}
	// find the chat settings relevant at this time
	flag := common.ChatFlag(0)
	if self.State == common.GameStateCreated {
		flag = self.ChatFlags[common.BeforeGamePhaseType]
	} else if self.State == common.GameStateEnded {
		flag = self.ChatFlags[common.AfterGamePhaseType]
	} else if phase != nil {
		flag = self.ChatFlags[phase.Type]
	} else {
		panic(fmt.Errorf("%+v has no phase??", self))
	}
	// if no chat is allowed, just return false
	if flag == 0 {
		return false
	}
	if message.Nation == common.Anonymous {
		// anonymous messages are only allowed if grey chat is allowed
		if (flag & common.ChatGrey) != common.ChatGrey {
			return false
		}
	} else if message.Nation != sender.Nation {
		// faked messages are only allowed if black chat is allowed
		if (flag & common.ChatBlack) != common.ChatBlack {
			return false
		}
	} else if (flag&common.ChatWhite) != common.ChatWhite && (flag&common.ChatBlack) != common.ChatBlack {
		// messages with the proper sender is only allowed if white or black chat is allowed
		return false
	}

	// if group chat is allowed, all recipient combinations are ok
	if (flag & common.ChatGroup) == common.ChatGroup {
		return true
	}

	// private messages are only allowed if private chat is allowed
	if len(message.Channel) == 2 {
		return (flag & common.ChatPrivate) == common.ChatPrivate
	}

	variant, found := common.VariantMap[self.Variant]
	if !found {
		panic(fmt.Errorf("Unknown variant for %+v", self))
	}

	// messages to everyone is only allowed if conference chat is allowed
	if len(message.Channel) == len(variant.Nations) {
		return (flag & common.ChatConference) == common.ChatConference
	}

	// otherwise, we dont allow it
	return false
}

func (self *Game) LastPhase(d *kol.DB) (result *Phase) {
	var phases Phases
	d.Query().Where(kol.Equals{"GameId", self.Id}).All(&phases)
	if len(phases) > 0 {
		sort.Sort(phases)
		result = &phases[0]
	}
	return
}

func (self *Game) Members(d *kol.DB) (result Members) {
	d.Query().Where(kol.Equals{"GameId", self.Id}).All(&result)
	sort.Sort(result)
	return
}

func (self *Game) Member(d *kol.DB, email string) (result *Member, err error) {
	var member Member
	var found bool
	if found, err = d.Query().Where(kol.And{kol.Equals{"GameId", self.Id}, kol.Equals{"UserId", kol.Id(email)}}).First(&member); found && err == nil {
		result = &member
	}
	return
}
