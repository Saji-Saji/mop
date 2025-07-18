package hunter

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
)

func (hunter *Hunter) registerExplosiveTrapSpell() {
	bonusPeriodicDamageMultiplier := core.TernaryFloat64(hunter.Spec == proto.Spec_SpecSurvivalHunter, 0.30, 0)
	cooldown := core.Ternary(hunter.Spec == proto.Spec_SpecSurvivalHunter, 24*time.Second, 30*time.Second)
	hunter.ExplosiveTrap = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 13812},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskSpellDamage,
		ClassSpellMask: HunterSpellExplosiveTrap,
		Flags:          core.SpellFlagAoE | core.SpellFlagAPL,

		FocusCost: core.FocusCostOptions{
			Cost: 0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: cooldown,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           hunter.CritMultiplier(1, 0),
		ThreatMultiplier:         1,

		Dot: core.DotConfig{
			IsAOE: true,
			Aura: core.Aura{
				Label: "Explosive Trap",
			},
			NumberOfTicks: 10,
			TickLength:    time.Second * 2,

			OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
				baseDamage := (27) + (0.0382 * dot.Spell.RangedAttackPower())
				dot.Spell.DamageMultiplierAdditive += bonusPeriodicDamageMultiplier
				dot.Spell.CalcAndDealPeriodicAoeDamage(sim, baseDamage, dot.Spell.OutcomeRangedHitAndCritNoBlock)
				dot.Spell.DamageMultiplierAdditive -= bonusPeriodicDamageMultiplier
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			if sim.CurrentTime < 0 {
				// Traps only last 60s.
				if sim.CurrentTime < -time.Second*60 {
					return
				}

				// If using this on prepull, the trap effect will go off when the fight starts
				// instead of immediately.
				pa := sim.GetConsumedPendingActionFromPool()

				pa.OnAction = func(sim *core.Simulation) {
					spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeRangedHitAndCritNoBlock, hunter.calcExplosiveTrapImpactDamage)
					hunter.ExplosiveTrap.AOEDot().Apply(sim)
				}

				sim.AddPendingAction(pa)
			} else {
				spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeRangedHitAndCritNoBlock, hunter.calcExplosiveTrapImpactDamage)
				hunter.ExplosiveTrap.AOEDot().Apply(sim)
			}
		},
	})
}

func (hunter *Hunter) calcExplosiveTrapImpactDamage(sim *core.Simulation, spell *core.Spell) float64 {
	baseDamage := (109 + sim.RandomFloat("Explosive Trap Initial")*125) + (0.0382 * spell.RangedAttackPower())
	baseDamage *= core.TernaryFloat64(hunter.Spec == proto.Spec_SpecSurvivalHunter, 1.3, 1)
	return baseDamage
}
