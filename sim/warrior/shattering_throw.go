package warrior

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
)

func (warrior *Warrior) RegisterShatteringThrowCD() {
	ShatteringThrowSpell := warrior.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 64382},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics,

		RageCost: core.RageCostOptions{
			Cost: 25 - float64(warrior.Talents.FocusedRage),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Minute * 5,
			},
			IgnoreHaste: true,
		},
		DamageMultiplier: 1,
		CritMultiplier:   warrior.critMultiplier(mh),
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 0.5 * spell.MeleeAttackPower()
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialNoBlockDodgeParry)
			if result.Landed() {
				core.ShatteringThrowAura(target).Activate(sim)
			}
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: ShatteringThrowSpell,
		Type:  core.CooldownTypeDPS,
		ActivationFactory: func(sim *core.Simulation) core.CooldownActivation {
			return func(sim *core.Simulation, character *core.Character) {
				if !warrior.StanceMatches(BattleStance) {
					if !warrior.BattleStance.IsReady(sim) {
						return
					}
					warrior.BattleStance.Cast(sim, nil)
				}
				if warrior.CurrentRage() < ShatteringThrowSpell.DefaultCast.Cost {
					return
				}
				if ShatteringThrowSpell.Cast(sim, character.CurrentTarget) {
					if warrior.AutoAttacks.MainhandSwingSpeed() == warrior.AutoAttacks.OffhandSwingSpeed() {
						warrior.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+ShatteringThrowSpell.CurCast.CastTime, true)
					} else {
						warrior.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+ShatteringThrowSpell.CurCast.CastTime, false)
					}
				}
			}
		},
		CanActivate: func(sim *core.Simulation, character *core.Character) bool {
			return warrior.CurrentRage() >= ShatteringThrowSpell.DefaultCast.Cost && (warrior.StanceMatches(BattleStance) || warrior.BattleStance.IsReady(sim))
		},
	})
}
