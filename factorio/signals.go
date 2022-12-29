package factorio

import "fmt"

type SignalType uint8

const (
	SignalTypeVirtual = iota
	SignalTypeItem
	SignalTypeFluid
)

func (t SignalType) MarshalJSON() ([]byte, error) {
	switch t {
	case SignalTypeVirtual:
		return []byte(`"virtual"`), nil
	case SignalTypeItem:
		return []byte(`"item"`), nil
	case SignalTypeFluid:
		return []byte(`"fluid"`), nil
	}

	return nil, fmt.Errorf("invalid signal type: %d", t)
}

// func (t *SignalType) UnmarshalJSON(raw []byte) error {
// 	str := string(raw)

// 	switch str {
// 	case "virtual":
// 		*t = SignalTypeVirtual
// 		return nil
// 	case "item":
// 		*t = SignalTypeItem
// 		return nil
// 	}

// 	return fmt.Errorf("unknown signal type: %s", str)
// }

type Signal struct {
	Type SignalType `json:"type"`
	Name string     `json:"name"`
}

var (
	// reserved
	signalEverything = &Signal{SignalTypeVirtual, "signal-everything"}
	signalEach       = &Signal{SignalTypeVirtual, "signal-each"}
	signalAny        = &Signal{SignalTypeVirtual, "signal-any"}
	SignalCheck      = &Signal{SignalTypeVirtual, "signal-check"}
	SignalG          = &Signal{SignalTypeVirtual, "signal-G"}
	SignalV          = &Signal{SignalTypeVirtual, "signal-V"}
	SignalS          = &Signal{SignalTypeVirtual, "signal-S"}

	Signal0       = &Signal{SignalTypeVirtual, "signal-0"}
	Signal1       = &Signal{SignalTypeVirtual, "signal-1"}
	Signal2       = &Signal{SignalTypeVirtual, "signal-2"}
	Signal3       = &Signal{SignalTypeVirtual, "signal-3"}
	Signal4       = &Signal{SignalTypeVirtual, "signal-4"}
	Signal5       = &Signal{SignalTypeVirtual, "signal-5"}
	Signal6       = &Signal{SignalTypeVirtual, "signal-6"}
	Signal7       = &Signal{SignalTypeVirtual, "signal-7"}
	Signal8       = &Signal{SignalTypeVirtual, "signal-8"}
	Signal9       = &Signal{SignalTypeVirtual, "signal-9"}
	SignalA       = &Signal{SignalTypeVirtual, "signal-A"}
	SignalB       = &Signal{SignalTypeVirtual, "signal-B"}
	SignalC       = &Signal{SignalTypeVirtual, "signal-C"}
	SignalD       = &Signal{SignalTypeVirtual, "signal-D"}
	SignalE       = &Signal{SignalTypeVirtual, "signal-E"}
	SignalF       = &Signal{SignalTypeVirtual, "signal-F"}
	SignalH       = &Signal{SignalTypeVirtual, "signal-H"}
	SignalI       = &Signal{SignalTypeVirtual, "signal-I"}
	SignalJ       = &Signal{SignalTypeVirtual, "signal-J"}
	SignalK       = &Signal{SignalTypeVirtual, "signal-K"}
	SignalL       = &Signal{SignalTypeVirtual, "signal-L"}
	SignalM       = &Signal{SignalTypeVirtual, "signal-M"}
	SignalN       = &Signal{SignalTypeVirtual, "signal-N"}
	SignalO       = &Signal{SignalTypeVirtual, "signal-O"}
	SignalP       = &Signal{SignalTypeVirtual, "signal-P"}
	SignalQ       = &Signal{SignalTypeVirtual, "signal-Q"}
	SignalR       = &Signal{SignalTypeVirtual, "signal-R"}
	SignalT       = &Signal{SignalTypeVirtual, "signal-T"}
	SignalU       = &Signal{SignalTypeVirtual, "signal-U"}
	SignalW       = &Signal{SignalTypeVirtual, "signal-W"}
	SignalX       = &Signal{SignalTypeVirtual, "signal-X"}
	SignalY       = &Signal{SignalTypeVirtual, "signal-Y"}
	SignalZ       = &Signal{SignalTypeVirtual, "signal-Z"}
	SignalBlack   = &Signal{SignalTypeVirtual, "signal-black"}
	SignalBlue    = &Signal{SignalTypeVirtual, "signal-blue"}
	SignalCyan    = &Signal{SignalTypeVirtual, "signal-cyan"}
	SignalDot     = &Signal{SignalTypeVirtual, "signal-dot"}
	SignalGreen   = &Signal{SignalTypeVirtual, "signal-green"}
	SignalGrey    = &Signal{SignalTypeVirtual, "signal-grey"}
	SignalInfo    = &Signal{SignalTypeVirtual, "signal-info"}
	SignalPink    = &Signal{SignalTypeVirtual, "signal-pink"}
	SignalRed     = &Signal{SignalTypeVirtual, "signal-red"}
	SignalUnknown = &Signal{SignalTypeVirtual, "signal-unknown"}
	SignalWhite   = &Signal{SignalTypeVirtual, "signal-white"}
	SignalYellow  = &Signal{SignalTypeVirtual, "signal-yellow"}

	SignalAccumulator                   = &Signal{SignalTypeItem, "accumulator"}
	SignalAdvancedCircuit               = &Signal{SignalTypeItem, "advanced-circuit"}
	SignalArithmeticCombinator          = &Signal{SignalTypeItem, "arithmetic-combinator"}
	SignalArtilleryTurret               = &Signal{SignalTypeItem, "artillery-turret"}
	SignalAssemblingMachine1            = &Signal{SignalTypeItem, "assembling-machine-1"}
	SignalAssemblingMachine2            = &Signal{SignalTypeItem, "assembling-machine-2"}
	SignalAssemblingMachine3            = &Signal{SignalTypeItem, "assembling-machine-3"}
	SignalBattery                       = &Signal{SignalTypeItem, "battery"}
	SignalBatteryEquipment              = &Signal{SignalTypeItem, "battery-equipment"}
	SignalBatteryMk2Equipment           = &Signal{SignalTypeItem, "battery-mk2-equipment"}
	SignalBeacon                        = &Signal{SignalTypeItem, "beacon"}
	SignalBeltImmunityEquipment         = &Signal{SignalTypeItem, "belt-immunity-equipment"}
	SignalBigElectricPole               = &Signal{SignalTypeItem, "big-electric-pole"}
	SignalBoiler                        = &Signal{SignalTypeItem, "boiler"}
	SignalBurnerGenerator               = &Signal{SignalTypeItem, "burner-generator"}
	SignalBurnerInserter                = &Signal{SignalTypeItem, "burner-inserter"}
	SignalBurnerMiningDrill             = &Signal{SignalTypeItem, "burner-mining-drill"}
	SignalCentrifuge                    = &Signal{SignalTypeItem, "centrifuge"}
	SignalChemicalPlant                 = &Signal{SignalTypeItem, "chemical-plant"}
	SignalCoal                          = &Signal{SignalTypeItem, "coal"}
	SignalCoin                          = &Signal{SignalTypeItem, "coin"}
	SignalConcrete                      = &Signal{SignalTypeItem, "concrete"}
	SignalConstantCombinator            = &Signal{SignalTypeItem, "constant-combinator"}
	SignalConstructionRobot             = &Signal{SignalTypeItem, "construction-robot"}
	SignalCopperCable                   = &Signal{SignalTypeItem, "copper-cable"}
	SignalCopperOre                     = &Signal{SignalTypeItem, "copper-ore"}
	SignalCopperPlate                   = &Signal{SignalTypeItem, "copper-plate"}
	SignalCrudeOilBarrel                = &Signal{SignalTypeItem, "crude-oil-barrel"}
	SignalDeciderCombinator             = &Signal{SignalTypeItem, "decider-combinator"}
	SignalDischargeDefenseEquipment     = &Signal{SignalTypeItem, "discharge-defense-equipment"}
	SignalElectricEnergyInterface       = &Signal{SignalTypeItem, "electric-energy-interface"}
	SignalElectricEngineUnit            = &Signal{SignalTypeItem, "electric-engine-unit"}
	SignalElectricFurnace               = &Signal{SignalTypeItem, "electric-furnace"}
	SignalElectricMiningDrill           = &Signal{SignalTypeItem, "electric-mining-drill"}
	SignalElectronicCircuit             = &Signal{SignalTypeItem, "electronic-circuit"}
	SignalEmptyBarrel                   = &Signal{SignalTypeItem, "empty-barrel"}
	SignalEnergyShieldEquipment         = &Signal{SignalTypeItem, "energy-shield-equipment"}
	SignalEnergyShieldMk2Equipment      = &Signal{SignalTypeItem, "energy-shield-mk2-equipment"}
	SignalEngineUnit                    = &Signal{SignalTypeItem, "engine-unit"}
	SignalExoskeletonEquipment          = &Signal{SignalTypeItem, "exoskeleton-equipment"}
	SignalExplosives                    = &Signal{SignalTypeItem, "explosives"}
	SignalExpressLoader                 = &Signal{SignalTypeItem, "express-loader"}
	SignalExpressSplitter               = &Signal{SignalTypeItem, "express-splitter"}
	SignalExpressTransportBelt          = &Signal{SignalTypeItem, "express-transport-belt"}
	SignalExpressUndergroundBelt        = &Signal{SignalTypeItem, "express-underground-belt"}
	SignalFastInserter                  = &Signal{SignalTypeItem, "fast-inserter"}
	SignalFastLoader                    = &Signal{SignalTypeItem, "fast-loader"}
	SignalFastSplitter                  = &Signal{SignalTypeItem, "fast-splitter"}
	SignalFastTransportBelt             = &Signal{SignalTypeItem, "fast-transport-belt"}
	SignalFastUndergroundBelt           = &Signal{SignalTypeItem, "fast-underground-belt"}
	SignalFilterInserter                = &Signal{SignalTypeItem, "filter-inserter"}
	SignalFlamethrowerTurret            = &Signal{SignalTypeItem, "flamethrower-turret"}
	SignalFlyingRobotFrame              = &Signal{SignalTypeItem, "flying-robot-frame"}
	SignalFusionReactorEquipment        = &Signal{SignalTypeItem, "fusion-reactor-equipment"}
	SignalGate                          = &Signal{SignalTypeItem, "gate"}
	SignalGreenWire                     = &Signal{SignalTypeItem, "green-wire"}
	SignalGunTurret                     = &Signal{SignalTypeItem, "gun-turret"}
	SignalHazardConcrete                = &Signal{SignalTypeItem, "hazard-concrete"}
	SignalHeatExchanger                 = &Signal{SignalTypeItem, "heat-exchanger"}
	SignalHeatInterface                 = &Signal{SignalTypeItem, "heat-interface"}
	SignalHeatPipe                      = &Signal{SignalTypeItem, "heat-pipe"}
	SignalHeavyOilBarrel                = &Signal{SignalTypeItem, "heavy-oil-barrel"}
	SignalInfinityChest                 = &Signal{SignalTypeItem, "infinity-chest"}
	SignalInfinityPipe                  = &Signal{SignalTypeItem, "infinity-pipe"}
	SignalInserter                      = &Signal{SignalTypeItem, "inserter"}
	SignalIronChest                     = &Signal{SignalTypeItem, "iron-chest"}
	SignalIronGearWheel                 = &Signal{SignalTypeItem, "iron-gear-wheel"}
	SignalIronOre                       = &Signal{SignalTypeItem, "iron-ore"}
	SignalIronPlate                     = &Signal{SignalTypeItem, "iron-plate"}
	SignalIronStick                     = &Signal{SignalTypeItem, "iron-stick"}
	SignalItemUnknown                   = &Signal{SignalTypeItem, "item-unknown"}
	SignalLab                           = &Signal{SignalTypeItem, "lab"}
	SignalLandMine                      = &Signal{SignalTypeItem, "land-mine"}
	SignalLandfill                      = &Signal{SignalTypeItem, "landfill"}
	SignalLaserTurret                   = &Signal{SignalTypeItem, "laser-turret"}
	SignalLightOilBarrel                = &Signal{SignalTypeItem, "light-oil-barrel"}
	SignalLinkedBelt                    = &Signal{SignalTypeItem, "linked-belt"}
	SignalLinkedChest                   = &Signal{SignalTypeItem, "linked-chest"}
	SignalLoader                        = &Signal{SignalTypeItem, "loader"}
	SignalLogisticChestActiveProvider   = &Signal{SignalTypeItem, "logistic-chest-active-provider"}
	SignalLogisticChestBuffer           = &Signal{SignalTypeItem, "logistic-chest-buffer"}
	SignalLogisticChestPassiveProvider  = &Signal{SignalTypeItem, "logistic-chest-passive-provider"}
	SignalLogisticChestRequester        = &Signal{SignalTypeItem, "logistic-chest-requester"}
	SignalLogisticChestStorage          = &Signal{SignalTypeItem, "logistic-chest-storage"}
	SignalLogisticRobot                 = &Signal{SignalTypeItem, "logistic-robot"}
	SignalLongHandedInserter            = &Signal{SignalTypeItem, "long-handed-inserter"}
	SignalLowDensityStructure           = &Signal{SignalTypeItem, "low-density-structure"}
	SignalLubricantBarrel               = &Signal{SignalTypeItem, "lubricant-barrel"}
	SignalMediumElectricPole            = &Signal{SignalTypeItem, "medium-electric-pole"}
	SignalNightVisionEquipment          = &Signal{SignalTypeItem, "night-vision-equipment"}
	SignalNuclearFuel                   = &Signal{SignalTypeItem, "nuclear-fuel"}
	SignalNuclearReactor                = &Signal{SignalTypeItem, "nuclear-reactor"}
	SignalOffshorePump                  = &Signal{SignalTypeItem, "offshore-pump"}
	SignalOilRefinery                   = &Signal{SignalTypeItem, "oil-refinery"}
	SignalPersonalLaserDefenseEquipment = &Signal{SignalTypeItem, "personal-laser-defense-equipment"}
	SignalPersonalRoboportEquipment     = &Signal{SignalTypeItem, "personal-roboport-equipment"}
	SignalPersonalRoboportMk2Equipment  = &Signal{SignalTypeItem, "personal-roboport-mk2-equipment"}
	SignalPetroleumGasBarrel            = &Signal{SignalTypeItem, "petroleum-gas-barrel"}
	SignalPipe                          = &Signal{SignalTypeItem, "pipe"}
	SignalPipeToGround                  = &Signal{SignalTypeItem, "pipe-to-ground"}
	SignalPlasticBar                    = &Signal{SignalTypeItem, "plastic-bar"}
	SignalPlayerPort                    = &Signal{SignalTypeItem, "player-port"}
	SignalPowerSwitch                   = &Signal{SignalTypeItem, "power-switch"}
	SignalProcessingUnit                = &Signal{SignalTypeItem, "processing-unit"}
	SignalProgrammableSpeaker           = &Signal{SignalTypeItem, "programmable-speaker"}
	SignalPump                          = &Signal{SignalTypeItem, "pump"}
	SignalPumpjack                      = &Signal{SignalTypeItem, "pumpjack"}
	SignalRadar                         = &Signal{SignalTypeItem, "radar"}
	SignalRailChainSignal               = &Signal{SignalTypeItem, "rail-chain-signal"}
	SignalRailSignal                    = &Signal{SignalTypeItem, "rail-signal"}
	SignalRedWire                       = &Signal{SignalTypeItem, "red-wire"}
	SignalRefinedConcrete               = &Signal{SignalTypeItem, "refined-concrete"}
	SignalRefinedHazardConcrete         = &Signal{SignalTypeItem, "refined-hazard-concrete"}
	SignalRoboport                      = &Signal{SignalTypeItem, "roboport"}
	SignalRocketControlUnit             = &Signal{SignalTypeItem, "rocket-control-unit"}
	SignalRocketFuel                    = &Signal{SignalTypeItem, "rocket-fuel"}
	SignalRocketPart                    = &Signal{SignalTypeItem, "rocket-part"}
	SignalRocketSilo                    = &Signal{SignalTypeItem, "rocket-silo"}
	SignalSatellite                     = &Signal{SignalTypeItem, "satellite"}
	SignalSimpleEntityWithForce         = &Signal{SignalTypeItem, "simple-entity-with-force"}
	SignalSimpleEntityWithOwner         = &Signal{SignalTypeItem, "simple-entity-with-owner"}
	SignalSmallElectricPole             = &Signal{SignalTypeItem, "small-electric-pole"}
	SignalSmallLamp                     = &Signal{SignalTypeItem, "small-lamp"}
	SignalSolarPanel                    = &Signal{SignalTypeItem, "solar-panel"}
	SignalSolarPanelEquipment           = &Signal{SignalTypeItem, "solar-panel-equipment"}
	SignalSolidFuel                     = &Signal{SignalTypeItem, "solid-fuel"}
	SignalSplitter                      = &Signal{SignalTypeItem, "splitter"}
	SignalStackFilterInserter           = &Signal{SignalTypeItem, "stack-filter-inserter"}
	SignalStackInserter                 = &Signal{SignalTypeItem, "stack-inserter"}
	SignalSteamEngine                   = &Signal{SignalTypeItem, "steam-engine"}
	SignalSteamTurbine                  = &Signal{SignalTypeItem, "steam-turbine"}
	SignalSteelChest                    = &Signal{SignalTypeItem, "steel-chest"}
	SignalSteelFurnace                  = &Signal{SignalTypeItem, "steel-furnace"}
	SignalSteelPlate                    = &Signal{SignalTypeItem, "steel-plate"}
	SignalStone                         = &Signal{SignalTypeItem, "stone"}
	SignalStoneBrick                    = &Signal{SignalTypeItem, "stone-brick"}
	SignalStoneFurnace                  = &Signal{SignalTypeItem, "stone-furnace"}
	SignalStoneWall                     = &Signal{SignalTypeItem, "stone-wall"}
	SignalStorageTank                   = &Signal{SignalTypeItem, "storage-tank"}
	SignalSubstation                    = &Signal{SignalTypeItem, "substation"}
	SignalSulfur                        = &Signal{SignalTypeItem, "sulfur"}
	SignalSulfuricAcidBarrel            = &Signal{SignalTypeItem, "sulfuric-acid-barrel"}
	SignalTrainStop                     = &Signal{SignalTypeItem, "train-stop"}
	SignalTransportBelt                 = &Signal{SignalTypeItem, "transport-belt"}
	SignalUndergroundBelt               = &Signal{SignalTypeItem, "underground-belt"}
	SignalUranium235                    = &Signal{SignalTypeItem, "uranium-235"}
	SignalUranium238                    = &Signal{SignalTypeItem, "uranium-238"}
	SignalUraniumFuelCell               = &Signal{SignalTypeItem, "uranium-fuel-cell"}
	SignalUraniumOre                    = &Signal{SignalTypeItem, "uranium-ore"}
	SignalUsedUpUraniumFuelCell         = &Signal{SignalTypeItem, "used-up-uranium-fuel-cell"}
	SignalWaterBarrel                   = &Signal{SignalTypeItem, "water-barrel"}
	SignalWood                          = &Signal{SignalTypeItem, "wood"}
	SignalWoodenChest                   = &Signal{SignalTypeItem, "wooden-chest"}
	SignalArtilleryShell                = &Signal{SignalTypeItem, "artillery-shell"}
	SignalAtomicBomb                    = &Signal{SignalTypeItem, "atomic-bomb"}
	SignalCannonShell                   = &Signal{SignalTypeItem, "cannon-shell"}
	SignalExplosiveCannonShell          = &Signal{SignalTypeItem, "explosive-cannon-shell"}
	SignalExplosiveRocket               = &Signal{SignalTypeItem, "explosive-rocket"}
	SignalExplosiveUraniumCannonShell   = &Signal{SignalTypeItem, "explosive-uranium-cannon-shell"}
	SignalFirearmMagazine               = &Signal{SignalTypeItem, "firearm-magazine"}
	SignalFlamethrowerAmmo              = &Signal{SignalTypeItem, "flamethrower-ammo"}
	SignalPiercingRoundsMagazine        = &Signal{SignalTypeItem, "piercing-rounds-magazine"}
	SignalPiercingShotgunShell          = &Signal{SignalTypeItem, "piercing-shotgun-shell"}
	SignalRocket                        = &Signal{SignalTypeItem, "rocket"}
	SignalShotgunShell                  = &Signal{SignalTypeItem, "shotgun-shell"}
	SignalUraniumCannonShell            = &Signal{SignalTypeItem, "uranium-cannon-shell"}
	SignalUraniumRoundsMagazine         = &Signal{SignalTypeItem, "uranium-rounds-magazine"}

	SignalCrudeOil     = &Signal{SignalTypeFluid, "crude-oil"}
	SignalFluidUnknown = &Signal{SignalTypeFluid, "fluid-unknown"}
	SignalHeavyOil     = &Signal{SignalTypeFluid, "heavy-oil"}
	SignalLightOil     = &Signal{SignalTypeFluid, "light-oil"}
	SignalLubricant    = &Signal{SignalTypeFluid, "lubricant"}
	SignalPetroleumGas = &Signal{SignalTypeFluid, "petroleum-gas"}
	SignalSteam        = &Signal{SignalTypeFluid, "steam"}
	SignalSulfuricAcid = &Signal{SignalTypeFluid, "sulfuric-acid"}
	SignalWater        = &Signal{SignalTypeFluid, "water"}
)

func (*Signal) isInput() {}

func (*Signal) isSecondaryInput() {}

var signals = [...]*Signal{
	Signal0,
	Signal1,
	Signal2,
	Signal3,
	Signal4,
	Signal5,
	Signal6,
	Signal7,
	Signal8,
	Signal9,
	SignalA,
	SignalB,
	SignalC,
	SignalD,
	SignalE,
	SignalF,
	SignalH,
	SignalI,
	SignalJ,
	SignalK,
	SignalL,
	SignalM,
	SignalN,
	SignalO,
	SignalP,
	SignalQ,
	SignalR,
	SignalT,
	SignalU,
	SignalW,
	SignalX,
	SignalY,
	SignalZ,
	SignalBlack,
	SignalBlue,
	SignalCyan,
	SignalDot,
	SignalGreen,
	SignalGrey,
	SignalInfo,
	SignalPink,
	SignalRed,
	SignalUnknown,
	SignalWhite,
	SignalYellow,
	SignalAccumulator,
	SignalAdvancedCircuit,
	SignalArithmeticCombinator,
	SignalArtilleryTurret,
	SignalAssemblingMachine1,
	SignalAssemblingMachine2,
	SignalAssemblingMachine3,
	SignalBattery,
	SignalBatteryEquipment,
	SignalBatteryMk2Equipment,
	SignalBeacon,
	SignalBeltImmunityEquipment,
	SignalBigElectricPole,
	SignalBoiler,
	SignalBurnerGenerator,
	SignalBurnerInserter,
	SignalBurnerMiningDrill,
	SignalCentrifuge,
	SignalChemicalPlant,
	SignalCoal,
	SignalCoin,
	SignalConcrete,
	SignalConstantCombinator,
	SignalConstructionRobot,
	SignalCopperCable,
	SignalCopperOre,
	SignalCopperPlate,
	SignalCrudeOilBarrel,
	SignalDeciderCombinator,
	SignalDischargeDefenseEquipment,
	SignalElectricEnergyInterface,
	SignalElectricEngineUnit,
	SignalElectricFurnace,
	SignalElectricMiningDrill,
	SignalElectronicCircuit,
	SignalEmptyBarrel,
	SignalEnergyShieldEquipment,
	SignalEnergyShieldMk2Equipment,
	SignalEngineUnit,
	SignalExoskeletonEquipment,
	SignalExplosives,
	SignalExpressLoader,
	SignalExpressSplitter,
	SignalExpressTransportBelt,
	SignalExpressUndergroundBelt,
	SignalFastInserter,
	SignalFastLoader,
	SignalFastSplitter,
	SignalFastTransportBelt,
	SignalFastUndergroundBelt,
	SignalFilterInserter,
	SignalFlamethrowerTurret,
	SignalFlyingRobotFrame,
	SignalFusionReactorEquipment,
	SignalGate,
	SignalGreenWire,
	SignalGunTurret,
	SignalHazardConcrete,
	SignalHeatExchanger,
	SignalHeatInterface,
	SignalHeatPipe,
	SignalHeavyOilBarrel,
	SignalInfinityChest,
	SignalInfinityPipe,
	SignalInserter,
	SignalIronChest,
	SignalIronGearWheel,
	SignalIronOre,
	SignalIronPlate,
	SignalIronStick,
	SignalItemUnknown,
	SignalLab,
	SignalLandMine,
	SignalLandfill,
	SignalLaserTurret,
	SignalLightOilBarrel,
	SignalLinkedBelt,
	SignalLinkedChest,
	SignalLoader,
	SignalLogisticChestActiveProvider,
	SignalLogisticChestBuffer,
	SignalLogisticChestPassiveProvider,
	SignalLogisticChestRequester,
	SignalLogisticChestStorage,
	SignalLogisticRobot,
	SignalLongHandedInserter,
	SignalLowDensityStructure,
	SignalLubricantBarrel,
	SignalMediumElectricPole,
	SignalNightVisionEquipment,
	SignalNuclearFuel,
	SignalNuclearReactor,
	SignalOffshorePump,
	SignalOilRefinery,
	SignalPersonalLaserDefenseEquipment,
	SignalPersonalRoboportEquipment,
	SignalPersonalRoboportMk2Equipment,
	SignalPetroleumGasBarrel,
	SignalPipe,
	SignalPipeToGround,
	SignalPlasticBar,
	SignalPlayerPort,
	SignalPowerSwitch,
	SignalProcessingUnit,
	SignalProgrammableSpeaker,
	SignalPump,
	SignalPumpjack,
	SignalRadar,
	SignalRailChainSignal,
	SignalRailSignal,
	SignalRedWire,
	SignalRefinedConcrete,
	SignalRefinedHazardConcrete,
	SignalRoboport,
	SignalRocketControlUnit,
	SignalRocketFuel,
	SignalRocketPart,
	SignalRocketSilo,
	SignalSatellite,
	SignalSimpleEntityWithForce,
	SignalSimpleEntityWithOwner,
	SignalSmallElectricPole,
	SignalSmallLamp,
	SignalSolarPanel,
	SignalSolarPanelEquipment,
	SignalSolidFuel,
	SignalSplitter,
	SignalStackFilterInserter,
	SignalStackInserter,
	SignalSteamEngine,
	SignalSteamTurbine,
	SignalSteelChest,
	SignalSteelFurnace,
	SignalSteelPlate,
	SignalStone,
	SignalStoneBrick,
	SignalStoneFurnace,
	SignalStoneWall,
	SignalStorageTank,
	SignalSubstation,
	SignalSulfur,
	SignalSulfuricAcidBarrel,
	SignalTrainStop,
	SignalTransportBelt,
	SignalUndergroundBelt,
	SignalUranium235,
	SignalUranium238,
	SignalUraniumFuelCell,
	SignalUraniumOre,
	SignalUsedUpUraniumFuelCell,
	SignalWaterBarrel,
	SignalWood,
	SignalWoodenChest,
	SignalArtilleryShell,
	SignalAtomicBomb,
	SignalCannonShell,
	SignalExplosiveCannonShell,
	SignalExplosiveRocket,
	SignalExplosiveUraniumCannonShell,
	SignalFirearmMagazine,
	SignalFlamethrowerAmmo,
	SignalPiercingRoundsMagazine,
	SignalPiercingShotgunShell,
	SignalRocket,
	SignalShotgunShell,
	SignalUraniumCannonShell,
	SignalUraniumRoundsMagazine,
	SignalCrudeOil,
	SignalFluidUnknown,
	SignalHeavyOil,
	SignalLightOil,
	SignalLubricant,
	SignalPetroleumGas,
	SignalSteam,
	SignalSulfuricAcid,
	SignalWater,
}
