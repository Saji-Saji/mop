package paladin

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

type Paladin struct {
	core.Character

	Seal      proto.PaladinSeal
	HolyPower HolyPowerBar

	Talents *proto.PaladinTalents

	// Used for CS/HotR
	sharedBuilderTimer *core.Timer

	CurrentSeal *core.Aura

	// Pets
	AncientGuardian *AncientGuardianPet

	AvengersShield *core.Spell
	Exorcism       *core.Spell
	HammerOfWrath  *core.Spell
	Judgment       *core.Spell

	AncientPowerAura        *core.Aura
	AvengingWrathAura       *core.Aura
	BastionOfGloryAura      *core.Aura
	BastionOfPowerAura      *core.Aura
	DivineCrusaderAura      *core.Aura
	DivineFavorAura         *core.Aura
	DivineProtectionAura    *core.Aura
	DivinePurposeAura       *core.Aura
	GoakAura                *core.Aura
	InfusionOfLightAura     *core.Aura
	SealOfInsightAura       *core.Aura
	SealOfJusticeAura       *core.Aura
	SealOfRighteousnessAura *core.Aura
	SealOfTruthAura         *core.Aura
	SelflessHealerAura      *core.Aura
	TheArtOfWarAura         *core.Aura

	// Item sets
	T11Ret4pc                *core.Aura
	T15Ret4pc                *core.Aura
	T15Ret4pcTemplarsVerdict *core.Spell

	HolyAvengerActionIDFilter  []core.ActionID
	JudgmentsOfTheWiseActionID core.ActionID
	DefensiveCooldownAuras     []*core.Aura

	DynamicHolyPowerSpent                        int32
	BastionOfGloryMultiplier                     float64
	ShieldOfTheRighteousAdditiveMultiplier       float64
	ShieldOfTheRighteousMultiplicativeMultiplier float64
}

// Implemented by each Paladin spec.
type PaladinAgent interface {
	GetPaladin() *Paladin
}

func (paladin *Paladin) GetCharacter() *core.Character {
	return &paladin.Character
}

func (paladin *Paladin) HasMajorGlyph(glyph proto.PaladinMajorGlyph) bool {
	return paladin.HasGlyph(int32(glyph))
}
func (paladin *Paladin) HasMinorGlyph(glyph proto.PaladinMinorGlyph) bool {
	return paladin.HasGlyph(int32(glyph))
}

func (paladin *Paladin) GetPaladin() *Paladin {
	return paladin
}

func (paladin *Paladin) AddRaidBuffs(_ *proto.RaidBuffs) {
}

func (paladin *Paladin) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (paladin *Paladin) Initialize() {
	paladin.registerGlyphs()
	paladin.registerSpells()
	paladin.addMistsPvpGloves()
}

func (paladin *Paladin) registerSpells() {
	paladin.registerAvengingWrath()
	paladin.registerCrusaderStrike()
	paladin.registerDevotionAura()
	paladin.registerDivineProtection()
	paladin.registerFlashOfLight()
	paladin.registerForbearance()
	paladin.registerGuardianOfAncientKings()
	paladin.registerHammerOfTheRighteous()
	paladin.registerHammerOfWrath()
	paladin.registerJudgment()
	paladin.registerLayOnHands()
	paladin.registerSanctityOfBattle()
	paladin.registerSealOfInsight()
	paladin.registerSealOfRighteousness()
	paladin.registerSealOfTruth()
	paladin.registerWordOfGlory()
}

func (paladin *Paladin) Reset(sim *core.Simulation) {
	switch paladin.Seal {
	case proto.PaladinSeal_Truth:
		paladin.CurrentSeal = paladin.SealOfTruthAura
		paladin.SealOfTruthAura.Activate(sim)
	case proto.PaladinSeal_Insight:
		paladin.CurrentSeal = paladin.SealOfInsightAura
		paladin.SealOfInsightAura.Activate(sim)
	case proto.PaladinSeal_Righteousness:
		paladin.CurrentSeal = paladin.SealOfRighteousnessAura
		paladin.SealOfRighteousnessAura.Activate(sim)
	case proto.PaladinSeal_Justice:
		paladin.CurrentSeal = paladin.SealOfJusticeAura
		paladin.SealOfJusticeAura.Activate(sim)
	}
}

func (paladin *Paladin) OnEncounterStart(sim *core.Simulation) {
}

func NewPaladin(character *core.Character, talentsStr string, options *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},
		Seal:      options.Seal,
	}

	core.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr)

	paladin.PseudoStats.CanParry = true

	paladin.EnableManaBar()
	paladin.HolyPower = HolyPowerBar{
		DefaultSecondaryResourceBarImpl: paladin.NewDefaultSecondaryResourceBar(core.SecondaryResourceConfig{
			Type: proto.SecondaryResourceType_SecondaryResourceTypeHolyPower,
			Max:  5,
		}),
		paladin: paladin,
	}
	paladin.RegisterSecondaryResourceBar(paladin.HolyPower)

	// Only retribution and holy are actually pets performing some kind of action
	if paladin.Spec != proto.Spec_SpecProtectionPaladin {
		paladin.AncientGuardian = paladin.NewAncientGuardian()
	}

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(paladin.DefaultCritMultiplier()),
		AutoSwingMelee: true,
	})

	paladin.PseudoStats.BaseBlockChance += 0.03
	paladin.PseudoStats.BaseDodgeChance += 0.03
	paladin.PseudoStats.BaseParryChance += 0.03

	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])

	// Base strength to Parry is not affected by Diminishing Returns
	baseStrength := paladin.GetBaseStats()[stats.Strength]
	paladin.PseudoStats.BaseParryChance += baseStrength * core.StrengthToParryPercent
	paladin.AddStat(stats.ParryRating, -baseStrength*core.StrengthToParryRating)
	paladin.AddStatDependency(stats.Strength, stats.ParryRating, core.StrengthToParryRating)
	paladin.AddStatDependency(stats.Agility, stats.DodgeRating, 0.1/10000.0/100.0)

	// Bonus Armor and Armor are treated identically for Paladins
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}

func (paladin *Paladin) CanTriggerHolyAvengerHpGain(actionID core.ActionID) {
	paladin.HolyAvengerActionIDFilter = append(paladin.HolyAvengerActionIDFilter, actionID)
}

// Shared cooldown for CS and HotR
func (paladin *Paladin) BuilderCooldown() *core.Timer {
	return paladin.Character.GetOrInitTimer(&paladin.sharedBuilderTimer)
}

func (paladin *Paladin) SpendableHolyPower() int32 {
	return min(paladin.HolyPower.Value(), 3)
}

func (paladin *Paladin) AddDefensiveCooldownAura(aura *core.Aura) {
	paladin.DefensiveCooldownAuras = append(paladin.DefensiveCooldownAuras, aura)
}

func (paladin *Paladin) AnyActiveDefensiveCooldown() bool {
	for _, aura := range paladin.DefensiveCooldownAuras {
		if aura.IsActive() {
			return true
		}
	}

	return false
}
