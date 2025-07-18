package protection

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
	"github.com/wowsims/mop/sim/paladin"
)

func RegisterProtectionPaladin() {
	core.RegisterAgentFactory(
		proto.Player_ProtectionPaladin{},
		proto.Spec_SpecProtectionPaladin,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewProtectionPaladin(character, options)
		},
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_ProtectionPaladin)
			if !ok {
				panic("Invalid spec value for Protection Paladin!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewProtectionPaladin(character *core.Character, options *proto.Player) *ProtectionPaladin {
	protOptions := options.GetProtectionPaladin()

	prot := &ProtectionPaladin{
		Paladin: paladin.NewPaladin(character, options.TalentsString, protOptions.Options.ClassOptions),
	}

	return prot
}

type ProtectionPaladin struct {
	*paladin.Paladin

	DamageTakenLastGlobal float64
}

func (prot *ProtectionPaladin) GetPaladin() *paladin.Paladin {
	return prot.Paladin
}

func (prot *ProtectionPaladin) Initialize() {
	prot.Paladin.Initialize()

	prot.registerMastery()

	prot.registerArdentDefender()
	prot.registerAvengersShieldSpell()
	prot.registerConsecrationSpell()
	prot.registerGrandCrusader()
	prot.registerGuardedByTheLight()
	prot.registerHolyWrath()
	prot.registerJudgmentsOfTheWise()
	prot.registerRighteousFury()
	prot.registerSanctuary()
	prot.registerShieldOfTheRighteous()

	// Vengeance
	prot.RegisterVengeance(84839, nil)

	prot.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ClassMask:  paladin.SpellMaskSealOfTruth | paladin.SpellMaskCensure,
		FloatValue: -0.8,
	})

	prot.trackDamageTakenLastGlobal()

	prot.registerHotfixPassive()
}

func (prot *ProtectionPaladin) trackDamageTakenLastGlobal() {
	prot.DamageTakenLastGlobal = 0.0

	core.MakePermanent(prot.GetOrRegisterAura(core.Aura{
		Label: "Damage Taken Last Global",

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			prot.DamageTakenLastGlobal = 0.0
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() {
				damageTaken := result.Damage
				prot.DamageTakenLastGlobal += damageTaken
				if sim.Log != nil {
					prot.Log(sim, "Damage Taken Last Global: %0.2f", prot.DamageTakenLastGlobal)
				}

				pa := sim.GetConsumedPendingActionFromPool()
				pa.NextActionAt = sim.CurrentTime + core.GCDDefault

				pa.OnAction = func(sim *core.Simulation) {
					prot.DamageTakenLastGlobal -= damageTaken

					if sim.Log != nil {
						prot.Log(sim, "Damage Taken Last Global: %0.2f", prot.DamageTakenLastGlobal)
					}
				}

				sim.AddPendingAction(pa)
			}
		},
	}))
}

func (prot *ProtectionPaladin) ApplyTalents() {
	prot.Paladin.ApplyTalents()
	prot.ApplyArmorSpecializationEffect(stats.Stamina, proto.ArmorType_ArmorTypePlate, 86525)
}

func (prot *ProtectionPaladin) Reset(sim *core.Simulation) {
	prot.Paladin.Reset(sim)
}

func (prot *ProtectionPaladin) OnEncounterStart(sim *core.Simulation) {
	prot.HolyPower.ResetBarTo(sim, 1)
	prot.Paladin.OnEncounterStart(sim)
}
