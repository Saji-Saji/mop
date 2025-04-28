package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/mop/sim"
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	_ "github.com/wowsims/mop/sim/encounters" // Needed for preset encounters.
	"github.com/wowsims/mop/tools"
	"github.com/wowsims/mop/tools/database"
	"github.com/wowsims/mop/tools/database/dbc"
)

func writeGzipFile(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip writer on top of the file writer
	gw := gzip.NewWriter(f)
	defer gw.Close()

	// Write the data to the gzip writer
	_, err = gw.Write(data)
	return err
}

// To do a full re-scrape, delete the previous output file first.
// go run ./tools/database/gen_db -outDir=assets -gen=atlasloot
// go run ./tools/database/gen_db -outDir=assets -gen=db

var outDir = flag.String("outDir", "assets", "Path to output directory for writing generated .go files.")
var genAsset = flag.String("gen", "", "Asset to generate. Valid values are 'db', 'atlasloot', 'wowhead-items', 'wowhead-spells', 'wowhead-itemdb', 'mop-items', and 'wago-db2-items'")
var dbPath = flag.String("dbPath", "./tools/database/wowsims.db", "Location of wowsims.db file from the DB2ToSqliteTool")

func main() {
	flag.Parse()

	database.DatabasePath = *dbPath

	if *outDir == "" {
		panic("outDir flag is required!")
	}

	dbDir := fmt.Sprintf("%s/database", *outDir)
	inputsDir := fmt.Sprintf("%s/db_inputs", *outDir)

	if *genAsset == "atlasloot" {
		db := database.ReadAtlasLootData()
		db.WriteJson(fmt.Sprintf("%s/atlasloot_db.json", inputsDir))
		return
	} else if *genAsset == "reforge-stats" {
		//Todo: fill this when we have information from wowhead @ Neteyes - Gehennas
		// For now, the version we have was taken from https://web.archive.org/web/20120201045249js_/http://www.wowhead.com/data=item-scaling
		return
	} else if *genAsset != "db" {
		panic("Invalid gen value")
	}
	helper, err := database.NewDBHelper()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer helper.Close()

	if err := database.RunOverrides(helper, "tools/database/overrides"); err != nil {
		log.Fatalf("failed to run overrides: %v", err)
	}

	database.GenerateProtos()

	randomSuffixes, err := database.LoadRawRandomSuffixes(helper)
	if err == nil {
		json, _ := json.Marshal(randomSuffixes)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/random_suffix.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	items, err := database.LoadRawItems(helper, "s.OverallQualityId != 7 AND s.Field_1_15_7_59706_054 = 0 AND s.OverallQualityId != 0 AND (i.ClassID = 2 OR i.ClassID = 4) AND s.Display_lang != '' AND (s.ID != 34219 AND s.Display_lang NOT LIKE '%Test%' AND s.Display_lang NOT LIKE 'QA%')")
	if err == nil {
		json, _ := json.Marshal(items)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/items.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	randPropsByIlvl, err := database.LoadRandomPropAllocations(helper)
	if err == nil {
		processed := make(dbc.RandomPropAllocationsByIlvl)
		for _, r := range randPropsByIlvl {
			processed[int(r.Ilvl)] = dbc.RandomPropAllocationMap{
				proto.ItemQuality_ItemQualityEpic:     [5]int32{r.Allocation.Epic0, r.Allocation.Epic1, r.Allocation.Epic2, r.Allocation.Epic3, r.Allocation.Epic4},
				proto.ItemQuality_ItemQualityRare:     [5]int32{r.Allocation.Superior0, r.Allocation.Superior1, r.Allocation.Superior2, r.Allocation.Superior3, r.Allocation.Superior4},
				proto.ItemQuality_ItemQualityUncommon: [5]int32{r.Allocation.Good0, r.Allocation.Good1, r.Allocation.Good2, r.Allocation.Good3, r.Allocation.Good4},
			}
		}
		json, _ := json.Marshal(processed)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/rand_prop_points.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	gems, err := database.LoadRawGems(helper)
	if err == nil {
		json, _ := json.Marshal(gems)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/gems.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	enchants, err := database.LoadRawEnchants(helper)
	if err == nil {
		json, _ := json.Marshal(enchants)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/enchants.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	spellEffects, err := database.LoadRawSpellEffects(helper)
	if err == nil {
		json, _ := json.Marshal(spellEffects)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/spell_effects.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	itemStatEffects, err := database.LoadItemStatEffects(helper)
	if err == nil {
		json, _ := json.Marshal(itemStatEffects)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_stat_effects.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	itemDamageTables, err := database.LoadItemDamageTables(helper)
	if err == nil {
		json, _ := json.Marshal(itemDamageTables)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_damage_tables.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	itemArmorTotal, err := database.LoadItemArmorTotal(helper)
	if err == nil {
		json, _ := json.Marshal(itemArmorTotal)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_armor_total.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	itemArmorQuality, err := database.LoadItemArmorQuality(helper)
	if err == nil {
		json, _ := json.Marshal(itemArmorQuality)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_armor_quality.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	} else {
		fmt.Println("Couldnt load quality")
	}

	itemArmorShield, err := database.LoadItemArmorShield(helper)
	if err == nil {
		json, _ := json.Marshal(itemArmorShield)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_armor_shield.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	armorLocation, err := database.LoadArmorLocation(helper)
	if err == nil {
		json, _ := json.Marshal(armorLocation)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/armor_location.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	itemeffects, err := database.LoadItemEffects(helper)
	if err == nil {
		json, _ := json.Marshal(itemeffects)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/item_effects.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	consumables, err := database.LoadConsumables(helper)
	if err == nil {
		json, _ := json.Marshal(consumables)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/consumables.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	}

	spellPowers, err := database.LoadSpellPowers(helper)
	if err == nil {
		json, _ := json.Marshal(spellPowers)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/spellpower.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	} else {
		log.Fatalf("Error %v", err)
	}

	spells, err := database.LoadSpells(helper)
	if err == nil {
		json, _ := json.Marshal(spells)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/spells.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	} else {
		log.Fatalf("Error %v", err)
	}
	dropSources, err := database.LoadDropSources(helper)
	if err == nil {
		json, _ := json.Marshal(dropSources)
		if err := writeGzipFile(fmt.Sprintf("%s/dbc/dropSources.json", inputsDir), json); err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
	} else {
		log.Fatalf("Error %v", err)
	}

	//Todo: See if we cant get rid of these as well
	atlaslootDB := database.ReadDatabaseFromJson(tools.ReadFile(fmt.Sprintf("%s/atlasloot_db.json", inputsDir)))

	// Todo: https://web.archive.org/web/20120201045249js_/http://www.wowhead.com/data=item-scaling
	reforgeStats := database.ParseWowheadReforgeStats(tools.ReadFile(fmt.Sprintf("%s/wowhead_reforge_stats.json", inputsDir)))

	db := database.NewWowDatabase()
	db.Encounters = core.PresetEncounters
	db.GlyphIDs = getGlyphIDsFromJson(fmt.Sprintf("%s/glyph_id_map.json", inputsDir))
	db.ReforgeStats = reforgeStats.ToProto()

	iconsMap, _ := database.LoadArtTexturePaths("./tools/DB2ToSqlite/listfile.csv")
	var instance = dbc.GetDBC()
	instance.LoadSpellScaling()

	for _, item := range instance.Items {
		parsed := item.ToUIItem()
		if parsed.Icon == "" {
			parsed.Icon = strings.ToLower(database.GetIconName(iconsMap, item.FDID))
		}
		db.MergeItem(parsed)
	}

	for _, gem := range instance.Gems {
		parsed := gem.ToProto()
		if parsed.Icon == "" {
			parsed.Icon = strings.ToLower(database.GetIconName(iconsMap, gem.FDID))
		}
		db.MergeGem(parsed)
	}

	for _, enchant := range instance.Enchants {
		parsed := enchant.ToProto()
		if parsed.Icon == "" {
			parsed.Icon = strings.ToLower(database.GetIconName(iconsMap, enchant.FDID))
		}
		db.MergeEnchant(parsed)
	}
	db.Spells = make(map[int32]*proto.Spell)
	for _, spell := range instance.Spells {

		// We want to preload all class spells
		if spell.SpellClassSet != 0 {
			db.Spells[spell.ID] = spell.ToProto()
		}

	}

	for _, item := range atlaslootDB.Items {
		if _, ok := db.Items[item.Id]; ok {
			db.MergeItem(item)
		}
	}
	for _, consumable := range consumables {
		protoConsumable := consumable.ToProto()
		protoConsumable.Icon = strings.ToLower(database.GetIconName(iconsMap, consumable.IconFileDataID))
		db.MergeConsumable(protoConsumable)
	}

	for _, consumable := range database.ConsumableOverrides {
		db.MergeConsumable(consumable)
	}
	db.MergeItems(database.ItemOverrides)
	db.MergeGems(database.GemOverrides)
	db.MergeEnchants(database.EnchantOverrides)

	ApplyGlobalFilters(db)

	leftovers := db.Clone()
	ApplyNonSimmableFilters(leftovers)
	leftovers.WriteBinaryAndJson(fmt.Sprintf("%s/leftover_db.bin", dbDir), fmt.Sprintf("%s/leftover_db.json", dbDir))

	ApplySimmableFilters(db)
	for _, enchant := range db.Enchants {
		if enchant.ItemId != 0 {
			db.AddItemIcon(enchant.ItemId, enchant.Icon, enchant.Name)
		}
		if enchant.SpellId != 0 {
			db.AddSpellIcon(enchant.ItemId, enchant.Icon, enchant.Name)
		}
	}

	for _, consume := range db.Consumables {
		if len(consume.EffectIds) > 0 {
			for _, se := range consume.EffectIds {
				effect := instance.SpellEffectsById[int(se)]
				db.MergeEffect(effect.ToProto())
			}
		}
	}

	for _, spell := range db.Spells {
		if len(spell.SpellEffects) > 0 {
			for _, se := range spell.SpellEffects {
				effect := instance.SpellEffectsById[int(se)]
				db.MergeEffect(effect.ToProto())
			}
		}
	}

	for _, randomSuffix := range dbc.GetDBC().RandomSuffix {
		if _, exists := db.RandomSuffixes[int32(randomSuffix.ID)]; !exists {
			db.RandomSuffixes[int32(randomSuffix.ID)] = randomSuffix.ToProto()
		}
	}

	icons, err := database.LoadSpellIcons(helper)
	for _, item := range db.Items {
		for _, source := range item.Sources {
			if crafted := source.GetCrafted(); crafted != nil {
				iconEntry := icons[int(crafted.SpellId)]
				icon := &proto.IconData{Id: int32(iconEntry.SpellID), Name: iconEntry.Name, Icon: strings.ToLower(database.GetIconName(iconsMap, iconEntry.FDID)), HasBuff: iconEntry.HasBuff}
				if iconEntry.SpellID == 0 {
					continue
				}
				db.SpellIcons[crafted.SpellId] = icon
			}
		}

		// Fetch Journal entry data
		dropSource := dropSources[int(item.Id)]
		if dropSource != nil {
			sources := make([]*proto.UIItemSource, 0, len(dropSource))
			for _, drop := range dropSource {
				sources = append(sources, &proto.UIItemSource{
					Source: &proto.UIItemSource_Drop{
						Drop: drop,
					},
				})
			}
			item.Sources = sources

		}
		// Auto-populate phase information if missing on Wowhead
		if item.Phase < 2 {
			item.Phase = InferPhase(item)
		}
	}

	if err != nil {
		panic("error loading icons")
	}
	for _, spellId := range database.SharedSpellsIcons {
		iconEntry := icons[int(spellId)]
		if iconEntry.Name == "" {
			continue
		}
		icon := &proto.IconData{Id: int32(iconEntry.SpellID), Name: iconEntry.Name, Icon: strings.ToLower(database.GetIconName(iconsMap, iconEntry.FDID)), HasBuff: iconEntry.HasBuff}
		if iconEntry.SpellID == 0 {
			continue
		}
		db.SpellIcons[spellId] = icon
	}

	for _, spellIds := range GetAllTalentSpellIds(&inputsDir) {
		for _, spellId := range spellIds {
			iconEntry := icons[int(spellId)]
			if iconEntry.Name == "" {
				continue
			}

			icon := &proto.IconData{Id: int32(iconEntry.SpellID), Name: iconEntry.Name, Icon: strings.ToLower(database.GetIconName(iconsMap, iconEntry.FDID)), HasBuff: iconEntry.HasBuff}
			if iconEntry.SpellID == 0 {
				continue
			}
			db.SpellIcons[spellId] = icon
		}
	}

	for _, spellIds := range GetAllRotationSpellIds() {
		for _, spellId := range spellIds {
			iconEntry := icons[int(spellId)]
			if iconEntry.Name == "" {
				continue
			}
			icon := &proto.IconData{Id: int32(iconEntry.SpellID), Name: iconEntry.Name, Icon: strings.ToLower(database.GetIconName(iconsMap, iconEntry.FDID)), HasBuff: iconEntry.HasBuff}
			if iconEntry.SpellID == 0 {
				continue
			}
			db.SpellIcons[spellId] = icon
		}
	}

	descriptions := make(map[int32]string)
	for _, enchant := range db.Enchants {
		var dbcEnch = instance.Enchants[int(enchant.EffectId)]
		descriptions[enchant.EffectId] = dbcEnch.EffectName
	}
	file, err := os.Create("assets/enchants/descriptions.json")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(descriptions); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	atlasDBProto := atlaslootDB.ToUIProto()
	db.MergeZones(atlasDBProto.Zones)
	db.MergeNpcs(atlasDBProto.Npcs)

	db.WriteBinaryAndJson(fmt.Sprintf("%s/db.bin", dbDir), fmt.Sprintf("%s/db.json", dbDir))
}

// Uses heuristics on ilvl + source to infer release phase of an item when missing.
func InferPhase(item *proto.UIItem) int32 {
	if item.Ilvl <= 463 {
		return 1
	}
	// Since we populate some drops before we run inferphase, we can use Atlasloot data to help us infer the phase
	for _, source := range item.Sources {
		if drop := source.GetDrop(); drop != nil {
			switch drop.ZoneId {
			case 6297, 6125, 6067: // MSV, Terrace, Heart of Fear
				return 1
			case 6622: // Throne of Thunder
				return 2
			case 6738: //Siege
				return 3
			}
		}
	}

	// if item.Ilvl >= 397 {
	// 	return 4 // Heroic Rag loot should already be tagged correctly by Wowhead.
	// }

	switch item.Ilvl {
	// case 353:
	// 	return 2
	// case 358, 371, 391:
	// 	return 3
	// case 359:
	// 	if item.Quality == proto.ItemQuality_ItemQualityUncommon {
	// 		return 4
	// 	}

	// 	return 1
	// case 372, 379:
	// 	return 1
	// case 377, 390:
	// 	return 4
	// case 365:
	// 	if strings.Contains(item.Name, "Vicious") {
	// 		return 1
	// 	}

	// 	return 3
	// case 378:
	// 	for _, itemSource := range item.Sources {
	// 		dropSource := itemSource.GetDrop()

	// 		if (dropSource != nil) && slices.Contains([]int32{5788, 5789, 5844}, dropSource.ZoneId) {
	// 			return 4
	// 		}
	// 	}

	// 	return 3
	// case 384:
	// 	if strings.Contains(item.Name, "Ruthless") {
	// 		return 3
	// 	}

	// 	return 4
	default:
		return 0
	}
}

// Filters out entities which shouldn't be included anywhere.
func ApplyGlobalFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		if _, ok := database.ItemAllowList[item.Id]; ok {
			return true
		}
		if _, ok := database.ItemDenyList[item.Id]; ok {
			return false
		}
		if item.Ilvl > 600 || item.Ilvl < 100 {
			return false
		}
		for _, pattern := range database.DenyListNameRegexes {
			if pattern.MatchString(item.Name) {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of every naxx set, e.g. https://www.wowhead.com/mop-classic/item=43728/bonescythe-gauntlets
	heroesItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasPrefix(item.Name, "Heroes' ")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := "Heroes' " + item.Name
		for _, heroItem := range heroesItems {
			if heroItem.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of many t8 set pieces, e.g. https://www.wowhead.com/mop-classic/item=46235/darkruned-gauntlets
	valorousItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasPrefix(item.Name, "Valorous ")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := "Valorous " + item.Name
		for _, item := range valorousItems {
			if item.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of many t9 set pieces, e.g. https://www.wowhead.com/mop-classic/item=48842/thralls-hauberk
	triumphItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasSuffix(item.Name, "of Triumph")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := item.Name + " of Triumph"
		for _, item := range triumphItems {
			if item.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	// Theres an invalid 251 t10 set for every class
	// The invalid set has a higher item id than the 'correct' ones
	t10invalidItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return item.SetName != "" && item.Ilvl == 251
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		for _, t10item := range t10invalidItems {
			if t10item.Name == item.Name && item.Ilvl == t10item.Ilvl && item.Id > t10item.Id {
				return false
			}
		}
		return true
	})

	db.Gems = core.FilterMap(db.Gems, func(_ int32, gem *proto.UIGem) bool {
		if _, ok := database.GemDenyList[gem.Id]; ok {
			return false
		}

		for _, pattern := range database.DenyListNameRegexes {
			if pattern.MatchString(gem.Name) {
				return false
			}
		}
		return true
	})

	db.Gems = core.FilterMap(db.Gems, func(_ int32, gem *proto.UIGem) bool {
		if strings.HasSuffix(gem.Name, "Stormjewel") {
			gem.Unique = false
		}
		return true
	})

	db.ItemIcons = core.FilterMap(db.ItemIcons, func(_ int32, icon *proto.IconData) bool {
		return icon.Name != "" && icon.Icon != ""
	})
	db.SpellIcons = core.FilterMap(db.SpellIcons, func(_ int32, icon *proto.IconData) bool {
		return icon.Name != "" && icon.Icon != "" && icon.Id != 0
	})

	db.Enchants = core.FilterMap(db.Enchants, func(_ database.EnchantDBKey, enchant *proto.UIEnchant) bool {
		for _, pattern := range database.DenyListNameRegexes {
			if pattern.MatchString(enchant.Name) {
				return false
			}
		}
		return !strings.HasPrefix(enchant.Name, "QA") && !strings.HasPrefix(enchant.Name, "Test") && !strings.HasPrefix(enchant.Name, "TEST")
	})

	db.Consumables = core.FilterMap(db.Consumables, func(_ int32, consumable *proto.Consumable) bool {
		if slices.Contains(database.ConsumableAllowList, consumable.Id) {
			return true
		}
		if slices.Contains(database.ConsumableDenyList, consumable.Id) {
			return false
		}

		if allZero(consumable.Stats) && consumable.Type != proto.ConsumableType_ConsumableTypePotion {
			return false
		}

		for _, pattern := range database.DenyListNameRegexes {
			if pattern.MatchString(consumable.Name) {
				return false
			}
		}

		if consumable.Type == proto.ConsumableType_ConsumableTypeUnknown || consumable.Type == proto.ConsumableType_ConsumableTypeScroll {
			return false
		}

		return !strings.HasPrefix(consumable.Name, "QA") && !strings.HasPrefix(consumable.Name, "Test") && !strings.HasPrefix(consumable.Name, "TEST")
	})
}

func allZero(stats []float64) bool {
	for _, val := range stats {
		if val != 0 {
			return false
		}
	}
	return true
}

// Filters out entities which shouldn't be included in the sim.
func ApplySimmableFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, simmableItemFilter)
	db.Gems = core.FilterMap(db.Gems, simmableGemFilter)
}

func ApplyNonSimmableFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, func(id int32, item *proto.UIItem) bool {
		return !simmableItemFilter(id, item)
	})
	db.Gems = core.FilterMap(db.Gems, func(id int32, gem *proto.UIGem) bool {
		return !simmableGemFilter(id, gem)
	})
}
func simmableItemFilter(_ int32, item *proto.UIItem) bool {
	if _, ok := database.ItemAllowList[item.Id]; ok {
		return true
	}

	if item.Quality < proto.ItemQuality_ItemQualityUncommon {
		return false
	} else if item.Quality == proto.ItemQuality_ItemQualityArtifact {
		return false
	} else if item.Quality >= proto.ItemQuality_ItemQualityHeirloom {
		return false
	} else if item.Quality <= proto.ItemQuality_ItemQualityEpic {
		if item.Ilvl < 372 {
			return false
		}
	} else {
		// Epic and legendary items might come from classic, so use a lower ilvl threshold.
		if item.Ilvl <= 359 {
			return false
		}
	}
	if item.Ilvl == 0 {
		fmt.Printf("Missing ilvl: %s\n", item.Name)
	}

	return true
}
func simmableGemFilter(_ int32, gem *proto.UIGem) bool {
	if _, ok := database.GemAllowList[gem.Id]; ok {
		return true
	}

	// Arbitrary to filter out old gems
	if gem.Id < 46000 {
		return false
	}

	return gem.Quality >= proto.ItemQuality_ItemQualityUncommon
}

type TalentConfig struct {
	FieldName string `json:"fieldName"`
	// Spell ID for each rank of this talent.
	// Omitted ranks will be inferred by incrementing from the last provided rank.
	SpellId int32 `json:"spellId"`
}

type TalentTreeConfig struct {
	BackgroundUrl string         `json:"backgroundUrl"`
	Talents       []TalentConfig `json:"talents"`
}

func getSpellIdsFromTalentJson(infile *string) []int32 {
	data, err := os.ReadFile(*infile)

	if err != nil {
		log.Fatalf("failed to load talent json file: %s", err)
	}

	var buf bytes.Buffer

	err = json.Compact(&buf, []byte(data))
	if err != nil {
		log.Fatalf("failed to compact json: %s", err)
	}

	var talentTree TalentTreeConfig

	err = json.Unmarshal(buf.Bytes(), &talentTree)
	if err != nil {
		log.Fatalf("failed to parse talent to json %s", err)
	}
	spellIds := make([]int32, 0)

	for _, talent := range talentTree.Talents {
		spellIds = append(spellIds, talent.SpellId)
	}

	return spellIds
}

func GetAllTalentSpellIds(inputsDir *string) map[string][]int32 {
	talentsDir := fmt.Sprintf("%s/../../ui/core/talents/trees", *inputsDir)
	specFiles := []string{
		"death_knight.json",
		"druid.json",
		"hunter.json",
		"mage.json",
		"paladin.json",
		"priest.json",
		"rogue.json",
		"shaman.json",
		"warlock.json",
		"warrior.json",
		"monk.json",
	}

	ret_db := make(map[string][]int32, 0)

	for _, specFile := range specFiles {
		specPath := fmt.Sprintf("%s/%s", talentsDir, specFile)
		ret_db[specFile[:len(specFile)-5]] = getSpellIdsFromTalentJson(&specPath)
	}

	return ret_db

}

type GlyphID struct {
	ItemID  int32 `json:"itemId"`
	SpellID int32 `json:"spellId"`
}

func getGlyphIDsFromJson(infile string) []*proto.GlyphID {
	data, err := os.ReadFile(infile)
	if err != nil {
		log.Fatalf("failed to load glyph json file: %s", err)
	}

	var buf bytes.Buffer
	err = json.Compact(&buf, []byte(data))
	if err != nil {
		log.Fatalf("failed to compact json: %s", err)
	}

	var glyphIDs []GlyphID

	err = json.Unmarshal(buf.Bytes(), &glyphIDs)
	if err != nil {
		log.Fatalf("failed to parse glyph IDs to json %s", err)
	}

	return core.MapSlice(glyphIDs, func(gid GlyphID) *proto.GlyphID {
		return &proto.GlyphID{
			ItemId:  gid.ItemID,
			SpellId: gid.SpellID,
		}
	})
}

func CreateTempAgent(r *proto.Raid) core.Agent {
	encounter := core.MakeSingleTargetEncounter(0.0)
	env, _, _ := core.NewEnvironment(r, encounter, false)
	return env.Raid.Parties[0].Players[0]
}

type RotContainer struct {
	Name string
	Raid *proto.Raid
}

func GetAllRotationSpellIds() map[string][]int32 {
	sim.RegisterAll()

	rotMapping := []RotContainer{
		// Death Knight
		{Name: "bloodDeathKnight", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDeathKnight,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_BloodDeathKnight{BloodDeathKnight: &proto.BloodDeathKnight{Options: &proto.BloodDeathKnight_Options{ClassOptions: &proto.DeathKnightOptions{}}, Rotation: &proto.BloodDeathKnight_Rotation{}}}), nil, nil, nil)},
		{Name: "frostDeathKnight", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDeathKnight,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_FrostDeathKnight{FrostDeathKnight: &proto.FrostDeathKnight{Options: &proto.FrostDeathKnight_Options{ClassOptions: &proto.DeathKnightOptions{}}}}), nil, nil, nil)},
		{Name: "unholyDeathKnight", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDeathKnight,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_UnholyDeathKnight{UnholyDeathKnight: &proto.UnholyDeathKnight{Options: &proto.UnholyDeathKnight_Options{ClassOptions: &proto.DeathKnightOptions{}}}}), nil, nil, nil)},

		// Druid
		{Name: "balanceDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},
		{Name: "feralDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_FeralDruid{FeralDruid: &proto.FeralDruid{Options: &proto.FeralDruid_Options{ClassOptions: &proto.DruidOptions{}}, Rotation: &proto.FeralDruid_Rotation{}}}), nil, nil, nil)},
		// {Name: "guardianDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
		// 	Class:         proto.Class_ClassDruid,
		// 	Equipment:     &proto.EquipmentSpec{},
		// 	TalentsString: "000000",
		// }, &proto.Player_FeralTankDruid{FeralTankDruid: &proto.FeralTankDruid{Options: &proto.FeralTankDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},
		{Name: "restorationDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_RestorationDruid{RestorationDruid: &proto.RestorationDruid{Options: &proto.RestorationDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},

		// Hunter
		{Name: "beastMasteryHunter", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassHunter,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_BeastMasteryHunter{BeastMasteryHunter: &proto.BeastMasteryHunter{Options: &proto.BeastMasteryHunter_Options{ClassOptions: &proto.HunterOptions{}}}}), nil, nil, nil)},
		{Name: "marksmanshipHunter", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassHunter,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_MarksmanshipHunter{MarksmanshipHunter: &proto.MarksmanshipHunter{Options: &proto.MarksmanshipHunter_Options{ClassOptions: &proto.HunterOptions{}}}}), nil, nil, nil)},
		{Name: "survivalHunter", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassHunter,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_SurvivalHunter{SurvivalHunter: &proto.SurvivalHunter{Options: &proto.SurvivalHunter_Options{ClassOptions: &proto.HunterOptions{}}}}), nil, nil, nil)},

		// Mage
		{Name: "arcaneMage", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMage,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ArcaneMage{ArcaneMage: &proto.ArcaneMage{Options: &proto.ArcaneMage_Options{ClassOptions: &proto.MageOptions{}}}}), nil, nil, nil)},
		{Name: "fireMage", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMage,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_FireMage{FireMage: &proto.FireMage{Options: &proto.FireMage_Options{ClassOptions: &proto.MageOptions{}}}}), nil, nil, nil)},
		{Name: "frostMage", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMage,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_FrostMage{FrostMage: &proto.FrostMage{Options: &proto.FrostMage_Options{ClassOptions: &proto.MageOptions{}}}}), nil, nil, nil)},

		// Paladin
		{Name: "holyPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{Options: &proto.HolyPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},
		{Name: "protPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ProtectionPaladin{ProtectionPaladin: &proto.ProtectionPaladin{Options: &proto.ProtectionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},
		{Name: "retPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Race:          proto.Race_RaceBloodElf,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},

		// Priest
		{Name: "disciplinePriest", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_DisciplinePriest{DisciplinePriest: &proto.DisciplinePriest{Options: &proto.DisciplinePriest_Options{ClassOptions: &proto.PriestOptions{}}}}), nil, nil, nil)},
		{Name: "holyPriest", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_HolyPriest{HolyPriest: &proto.HolyPriest{Options: &proto.HolyPriest_Options{ClassOptions: &proto.PriestOptions{}}}}), nil, nil, nil)},
		{Name: "shadowPriest", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{Options: &proto.ShadowPriest_Options{ClassOptions: &proto.PriestOptions{}}}}), nil, nil, nil)},

		// Rogue
		{Name: "assassinationRogue", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_AssassinationRogue{AssassinationRogue: &proto.AssassinationRogue{Options: &proto.AssassinationRogue_Options{ClassOptions: &proto.RogueOptions{}}}}), nil, nil, nil)},
		{Name: "combatRogue", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_CombatRogue{CombatRogue: &proto.CombatRogue{Options: &proto.CombatRogue_Options{ClassOptions: &proto.RogueOptions{}}}}), nil, nil, nil)},
		{Name: "subtletyRogue", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_SubtletyRogue{SubtletyRogue: &proto.SubtletyRogue{Options: &proto.SubtletyRogue_Options{ClassOptions: &proto.RogueOptions{}}}}), nil, nil, nil)},

		// Shaman
		{Name: "elementalShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},
		{Name: "enhancementShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},
		{Name: "restorationShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_RestorationShaman{RestorationShaman: &proto.RestorationShaman{Options: &proto.RestorationShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},

		// Warlock
		{Name: "afflictionWarlock", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarlock,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_AfflictionWarlock{AfflictionWarlock: &proto.AfflictionWarlock{Options: &proto.AfflictionWarlock_Options{ClassOptions: &proto.WarlockOptions{}}}}), nil, nil, nil)},
		{Name: "demonologyWarlock", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarlock,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_DemonologyWarlock{DemonologyWarlock: &proto.DemonologyWarlock{Options: &proto.DemonologyWarlock_Options{ClassOptions: &proto.WarlockOptions{}}}}), nil, nil, nil)},
		{Name: "destructionWarlock", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarlock,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_DestructionWarlock{DestructionWarlock: &proto.DestructionWarlock{Options: &proto.DestructionWarlock_Options{ClassOptions: &proto.WarlockOptions{}}}}), nil, nil, nil)},

		// Warrior
		{Name: "armsWarrior", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ArmsWarrior{ArmsWarrior: &proto.ArmsWarrior{Options: &proto.ArmsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}), nil, nil, nil)},
		{Name: "furyWarrior", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_FuryWarrior{FuryWarrior: &proto.FuryWarrior{Options: &proto.FuryWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}), nil, nil, nil)},
		{Name: "protectionWarrior", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{Options: &proto.ProtectionWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}), nil, nil, nil)},

		// Monk
		{Name: "brewmasterMonk", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMonk,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_BrewmasterMonk{BrewmasterMonk: &proto.BrewmasterMonk{Options: &proto.BrewmasterMonk_Options{ClassOptions: &proto.MonkOptions{}, Stance: proto.MonkStance_SturdyOx}}}), nil, nil, nil)},
		{Name: "mistweaverMonk", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMonk,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_MistweaverMonk{MistweaverMonk: &proto.MistweaverMonk{Options: &proto.MistweaverMonk_Options{ClassOptions: &proto.MonkOptions{}, Stance: proto.MonkStance_WiseSerpent}}}), nil, nil, nil)},
		{Name: "windwalkerMonk", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMonk,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: "000000",
		}, &proto.Player_WindwalkerMonk{WindwalkerMonk: &proto.WindwalkerMonk{Options: &proto.WindwalkerMonk_Options{ClassOptions: &proto.MonkOptions{}}}}), nil, nil, nil)},
	}

	ret_db := make(map[string][]int32, 0)

	for _, r := range rotMapping {
		f := CreateTempAgent(r.Raid).GetCharacter()

		spells := make([]int32, 0, len(f.Spellbook))

		for _, s := range f.Spellbook {
			if s.SpellID != 0 {
				spells = append(spells, s.SpellID)
			}
		}

		for _, s := range f.GetAuras() {
			if s.ActionID.SpellID != 0 {
				spells = append(spells, s.ActionID.SpellID)
			}
		}

		ret_db[r.Name] = spells
	}
	return ret_db
}
