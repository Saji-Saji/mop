import { Party } from '../party.js';
import { Player } from '../player';
import { ConsumesSpec, Debuffs, Faction, IndividualBuffs, PartyBuffs, RaidBuffs, Spec } from '../proto/common.js';
import { ActionId } from '../proto_utils/action_id.js';
import { Raid } from '../raid';
import { EventID, TypedEvent } from '../typed_event';
import * as InputHelpers from './input_helpers';
import { IconEnumPicker } from './pickers/icon_enum_picker.jsx';
import { IconPicker } from './pickers/icon_picker.jsx';

// Component Functions

export type IconInputConfig<ModObject, T> = InputHelpers.TypedIconPickerConfig<ModObject, T> | InputHelpers.TypedIconEnumPickerConfig<ModObject, T>;

export const buildIconInput = <SpecType extends Spec>(parent: HTMLElement, player: Player<SpecType>, inputConfig: IconInputConfig<Player<SpecType>, any>) => {
	if (inputConfig.type == 'icon') {
		return new IconPicker<Player<SpecType>, any>(parent, player, inputConfig);
	} else if (inputConfig.type == 'iconEnum') {
		return new IconEnumPicker<Player<SpecType>, any>(parent, player, inputConfig);
	} else {
		throw new Error('Unsupported input type');
	}
};

export function withLabel<ModObject, T>(config: IconInputConfig<ModObject, T>, label: string): IconInputConfig<ModObject, T> {
	config.label = label;
	return config;
}

interface BooleanInputConfig<T> {
	actionId: ActionId;
	fieldName: keyof T;
	value?: number;
	label?: string;
	faction?: Faction;
	showWhen?: (player: Player<any>) => boolean;
}

export function makeBooleanRaidBuffInput<SpecType extends Spec>(
	config: BooleanInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> {
	return InputHelpers.makeBooleanIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.getRaid()!.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
}
export function makeBooleanPartyBuffInput<SpecType extends Spec>(
	config: BooleanInputConfig<PartyBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> {
	return InputHelpers.makeBooleanIconInput<any, PartyBuffs, Party>(
		{
			getModObject: (player: Player<SpecType>) => player.getParty()!,
			getValue: (party: Party) => party.getBuffs(),
			setValue: (eventID: EventID, party: Party, newVal: PartyBuffs) => party.setBuffs(eventID, newVal),
			changeEmitter: (party: Party) => party.buffsChangeEmitter,
		},
		config.actionId,
		config.fieldName,
		config.value,
	);
}

export function makeBooleanIndividualBuffInput<SpecType extends Spec>(
	config: BooleanInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> {
	return InputHelpers.makeBooleanIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
}

export function makeBooleanConsumeInput<SpecType extends Spec>(
	config: BooleanInputConfig<ConsumesSpec>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> {
	return InputHelpers.makeBooleanIconInput<any, ConsumesSpec, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			getValue: (player: Player<SpecType>) => player.getConsumes(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: ConsumesSpec) => player.setConsumes(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.consumesChangeEmitter, player.professionChangeEmitter]),
			showWhen: (player: Player<SpecType>) => !config.showWhen || config.showWhen(player),
		},
		config.actionId,
		config.fieldName,
		config.value,
	);
}
export function makeBooleanDebuffInput<SpecType extends Spec>(
	config: BooleanInputConfig<Debuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> {
	return InputHelpers.makeBooleanIconInput<any, Debuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			getValue: (player: Player<SpecType>) => player.getRaid()!.getDebuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: Debuffs) => player.getRaid()!.setDebuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.getRaid()!.debuffsChangeEmitter]),
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
}

interface TristateInputConfig<T> {
	actionId: ActionId;
	impId: ActionId;
	fieldName: keyof T;
	faction?: Faction;
	label?: string;
}

export function makeTristateRaidBuffInput<SpecType extends Spec>(
	config: TristateInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeTristateIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.getRaid()!.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
}

export function makeTristateIndividualBuffInput<SpecType extends Spec>(
	config: TristateInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeTristateIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
}

export function makeTristateDebuffInput<SpecType extends Spec>(
	config: TristateInputConfig<Debuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeTristateIconInput<any, Debuffs, Raid>(
		{
			getModObject: (player: Player<SpecType>) => player.getRaid()!,
			getValue: (raid: Raid) => raid.getDebuffs(),
			setValue: (eventID: EventID, raid: Raid, newVal: Debuffs) => raid.setDebuffs(eventID, newVal),
			changeEmitter: (raid: Raid) => raid.debuffsChangeEmitter,
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
}

interface QuadStateInputConfig<T> {
	actionId: ActionId;
	impId: ActionId;
	impId2: ActionId;
	fieldName: keyof T;
	faction?: Faction;
}

export function makeQuadstateDebuffInput<SpecType extends Spec>(
	config: QuadStateInputConfig<Debuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeQuadstateIconInput<any, Debuffs, Raid>(
		{
			getModObject: (player: Player<SpecType>) => player.getRaid()!,
			getValue: (raid: Raid) => raid.getDebuffs(),
			setValue: (eventID: EventID, raid: Raid, newVal: Debuffs) => raid.setDebuffs(eventID, newVal),
			changeEmitter: (raid: Raid) => raid.debuffsChangeEmitter,
		},
		config.actionId,
		config.impId,
		config.impId2,
		config.fieldName,
	);
}

interface MultiStateInputConfig<T> {
	actionId: ActionId;
	label?: string;
	numStates: number;
	fieldName: keyof T;
	multiplier?: number;
	faction?: Faction;
}

export function makeMultistateRaidBuffInput<SpecType extends Spec>(
	config: MultiStateInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeMultistateIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.getRaid()!.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.numStates,
		config.fieldName,
		config.multiplier,
		config.label,
	);
}
export function makeMultistatePartyBuffInput<SpecType extends Spec>(
	actionId: ActionId,
	numStates: number,
	fieldName: keyof PartyBuffs,
	label?: string,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeMultistateIconInput<any, PartyBuffs, Party>(
		{
			getModObject: (player: Player<SpecType>) => player.getParty()!,
			getValue: (party: Party) => party.getBuffs(),
			setValue: (eventID: EventID, party: Party, newVal: PartyBuffs) => party.setBuffs(eventID, newVal),
			changeEmitter: (party: Party) => party.buffsChangeEmitter,
		},
		actionId,
		numStates,
		fieldName,
		undefined,
		label,
	);
}
export function makeMultistateIndividualBuffInput<SpecType extends Spec>(
	config: MultiStateInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeMultistateIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.buffsChangeEmitter, player.raceChangeEmitter]),
		},
		config.actionId,
		config.numStates,
		config.fieldName,
		config.multiplier,
		config.label,
	);
}

export function makeMultistateMultiplierIndividualBuffInput<SpecType extends Spec>(
	actionId: ActionId,
	numStates: number,
	multiplier: number,
	fieldName: keyof IndividualBuffs,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> {
	return InputHelpers.makeMultistateIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (eventID: EventID, player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(eventID, newVal),
			changeEmitter: (player: Player<SpecType>) => player.buffsChangeEmitter,
		},
		actionId,
		numStates,
		fieldName,
		multiplier,
	);
}

export function makeMultistateMultiplierDebuffInput<SpecType extends Spec>(
	actionId: ActionId,
	numStates: number,
	multiplier: number,
	fieldName: keyof Debuffs,
): InputHelpers.TypedIconPickerConfig<Player<any>, number> {
	return InputHelpers.makeMultistateIconInput<any, Debuffs, Raid>(
		{
			getModObject: (player: Player<SpecType>) => player.getRaid()!,
			getValue: (raid: Raid) => raid.getDebuffs(),
			setValue: (eventID: EventID, raid: Raid, newVal: Debuffs) => raid.setDebuffs(eventID, newVal),
			changeEmitter: (raid: Raid) => raid.debuffsChangeEmitter,
		},
		actionId,
		numStates,
		fieldName,
		multiplier,
	);
}

// interface EnumInputConfig<ModObject, Message, T> {
// 	fieldName: keyof Message
// 	values: Array<IconEnumValueConfig<ModObject, T>>
// 	direction?: IconEnumPickerDirection
// 	numColumns?: number
// 	faction?: Faction
// }

// export function makeEnumIndividualBuffInput<SpecType extends Spec>(config: EnumInputConfig<Player<SpecType>, IndividualBuffs, number>): InputHelpers.TypedIconEnumPickerConfig<Player<SpecType>, number> {
// 	return InputHelpers.makeEnumIconInput<any, IndividualBuffs, Player<SpecType>, number>({
// 		getModObject: (player: Player<SpecType>) => player,
// 		showWhen: (player: Player<SpecType>) =>
// 			(!config.faction || config.faction == player.getFaction()),
// 		getValue: (player: Player<SpecType>) => player.getBuffs(),
// 		setValue: (eventID: EventID, player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(eventID, newVal),
// 		changeEmitter: (player: Player<SpecType>) => TypedEvent.onAny([player.buffsChangeEmitter, player.raceChangeEmitter]),
// 	}, config.fieldName, config.values, config.numColumns, config.direction || IconEnumPickerDirection.Vertical)
// };
