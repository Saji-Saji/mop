package protection

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
	"github.com/wowsims/mop/sim/warrior"
)

func RegisterProtectionWarrior() {
	core.RegisterAgentFactory(
		proto.Player_ProtectionWarrior{},
		proto.Spec_SpecProtectionWarrior,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewProtectionWarrior(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_ProtectionWarrior)
			if !ok {
				panic("Invalid spec value for Protection Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

type ProtectionWarrior struct {
	*warrior.Warrior

	Options *proto.ProtectionWarrior_Options

	SwordAndBoardAura *core.Aura
}

func NewProtectionWarrior(character *core.Character, options *proto.Player) *ProtectionWarrior {
	protOptions := options.GetProtectionWarrior().Options

	war := &ProtectionWarrior{
		Warrior: warrior.NewWarrior(character, protOptions.ClassOptions, options.TalentsString, warrior.WarriorInputs{}),
		Options: protOptions,
	}

	return war
}

func (war *ProtectionWarrior) CalculateMasteryBlockChance(masteryRating float64) float64 {
	return 0.5 * (8.0 + core.MasteryRatingToMasteryPoints(masteryRating))
}

func (war *ProtectionWarrior) CalculateMasteryCriticalBlockChance() float64 {
	return 2.2 * (8.0 + war.GetMasteryPoints()) / 100.0
}

func (war *ProtectionWarrior) GetWarrior() *warrior.Warrior {
	return war.Warrior
}

func (war *ProtectionWarrior) Initialize() {
	war.Warrior.Initialize()
	war.registerPassives()

	war.registerDevastate()
	war.registerRevenge()
	war.registerShieldSlam()
	war.registerShieldBlock()
	war.registerShieldBarrier()
	war.registerDemoralizingShout()
	war.registerLastStand()
}

func (war *ProtectionWarrior) registerPassives() {
	war.ApplyArmorSpecializationEffect(stats.Stamina, proto.ArmorType_ArmorTypePlate, 86526)

	// Critical block
	war.registerMastery()

	war.registerUnwaveringSentinel()
	war.registerBastionOfDefense()
	war.registerSwordAndBoard()
	war.registerUltimatum()
	war.registerRiposte()

	// Vengeance
	war.RegisterVengeance(93098, war.DefensiveStanceAura)
}

func (war *ProtectionWarrior) registerMastery() {

	dummyCriticalBlockSpell := war.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 76857}, // Doesn't seem like there's an actual spell ID for the block itself, so use the mastery ID
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    war.NewTimer(),
				Duration: time.Second * 3,
			},
		},
	})

	war.Blockhandler = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		procChance := war.GetCriticalBlockChance()
		if dummyCriticalBlockSpell.CD.IsReady(sim) && sim.Proc(procChance, "Critical Block Roll") {
			result.Damage = result.Damage * (1 - war.BlockDamageReduction()*2)
			dummyCriticalBlockSpell.Cast(sim, spell.Unit)
			return
		}
		result.Damage = result.Damage * (1 - war.BlockDamageReduction())
	}

	war.CriticalBlockChance[0] = war.CalculateMasteryCriticalBlockChance()
	war.AddStat(stats.BlockPercent, war.CalculateMasteryBlockChance(war.GetStat(stats.MasteryRating)))

	war.AddOnMasteryStatChanged(func(sim *core.Simulation, oldMasteryRating float64, newMasteryRating float64) {
		masteryBlockStat := war.CalculateMasteryBlockChance(newMasteryRating - oldMasteryRating)
		war.AddStatDynamic(sim, stats.BlockPercent, masteryBlockStat)
		war.CriticalBlockChance[0] = war.CalculateMasteryCriticalBlockChance()
	})
}

func (war *ProtectionWarrior) Reset(sim *core.Simulation) {
	war.Warrior.Reset(sim)
}

func (war *ProtectionWarrior) OnEncounterStart(sim *core.Simulation) {
	war.ResetRageBar(sim, core.TernaryFloat64(war.ShieldBarrierAura.IsActive(), 5, 25)+war.PrePullChargeGain)
	war.Warrior.OnEncounterStart(sim)
}
