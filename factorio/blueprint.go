package factorio

import (
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"main/inspector"
	"main/util"
	"math"
	"strings"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type RingArea int

const (
	// Horizontal
	RingAreaTop RingArea = iota
	RingAreaBottom

	// Vertical
	RingAreaLeft
	RingAreaRight

	// Corners
	RingAreaTopLeft
	RingAreaTopRight
	RingAreaBottomLeft
	RingAreaBottomRight
)

func (p Position) isZero() bool {
	return p.X == 0 && p.Y == 0
}

func (p Position) InspectCustom() inspector.InspectString {
	return inspector.InspectString(fmt.Sprintf("(%.2f, %.2f)", p.X, p.Y))
}

func (p Position) getRingArea() (area RingArea, radius float64) {
	x, y := p.abs()

	if x == y {
		switch {
		case p.X < 0 && p.Y < 0:
			return RingAreaBottomLeft, x
		case p.X > 0 && p.Y < 0:
			return RingAreaBottomRight, x
		case p.X < 0 && p.Y > 0:
			return RingAreaTopLeft, x
		case p.X > 0 && p.Y > 0:
			return RingAreaTopRight, x
		}
	}

	if x > y {
		if p.X < 0 {
			return RingAreaLeft, x
		}

		return RingAreaRight, x
	}

	if p.Y < 0 {
		return RingAreaBottom, y
	}

	return RingAreaTop, y
}

type Plot struct {
	rings    []Ring
	entities []Entity
}

func (p Plot) InspectCustom() inspector.InspectString {
	if len(p.rings) == 0 {
		return "<empty plot>"
	}

	mr := p.rings[len(p.rings)-1]

	dim := (len(mr)/4 + 1)

	var b strings.Builder
	b.Grow(dim*dim + dim)
	rad := mr.radius()

	for y := -rad + .5; y < rad-.5; y++ {
		for x := -rad + .5; x < rad-.5; x++ {
			ent := p.get(Position{x, y})

			if ent == nil {
				b.WriteString("\x1b[90m.\x1b[0m")
				continue
			}

			switch ent.(type) {
			case *SubstationEntity:
				b.WriteString("\x1b[34mS\x1b[0m")
			case *ArithmeticCombinatorEntity:
				b.WriteString("\x1b[36ma\x1b[0m")
			case *DeciderCombinatorEntity:
				b.WriteString("\x1b[33md\x1b[0m")
			case *ConstantCombinatorEntity:
				b.WriteByte('c')
			case *LampEntity:
				b.WriteByte('l')
			case *RoboportEntity:
				b.WriteString("\x1b[35mR\x1b[0m")
			default:
				b.WriteString("\x1b[41m?\x1b[0m")
			}
		}

		if y != rad {
			b.WriteByte('\n')
		}
	}

	return inspector.InspectString(b.String())
}

func (p Plot) canPlace(b Bounds) (canPlace bool) {
	canPlace = true
	b.iterate(func(pos Position) (shouldBreak bool) {
		if p.get(pos) != nil {
			canPlace = false

			shouldBreak = true
		}

		return
	})

	return
}

func (p *Plot) remove(e Entity) {
	e.Bounds().iterate(func(pos Position) bool {
		ptr := p.getPtr(pos)

		if util.DebugEnabled && *ptr == nil {
			fmt.Println("entity missing", inspector.Inspect(pos))
		}

		*ptr = nil

		return false
	})

	b := e.basic()

	b.Position = Position{}
	b.Direction = DirectionDefault
}

func (p *Plot) place(e Entity, pos Position, d Direction) {
	// ensure entity has a valid number
	e.Number()
	b := e.basic()
	b.Position = pos
	b.Direction = d

	bounds := e.Bounds()

	ringIdx := bounds.BL.toRingIdx()

	if trIdx := bounds.TR.toRingIdx(); trIdx > ringIdx {
		ringIdx = trIdx
	}

	if ringIdx >= len(p.rings) {
		v, ok := e.(*SubstationEntity)
		fmt.Println(inspector.Inspect(v), inspector.Inspect(e))

		fmt.Println("ring idx out of bounds", len(p.rings), inspector.Inspect(d), ok)

		fmt.Printf("BL: %s; Ring index: %s\n", inspector.Inspect(bounds.BL), inspector.Inspect(bounds.BL.toRingIdx()))
		fmt.Printf("TR: %s; Ring index: %s\n", inspector.Inspect(bounds.TR), inspector.Inspect(bounds.TR.toRingIdx()))
		inspector.Println(e)
		inspector.Println(p)
		// panic("unable to place")
	}

	p.entities = append(p.entities, e)

	bounds.iterate(func(pos Position) (shouldBreak bool) {
		ptr := p.getPtr(pos)

		if util.DebugEnabled && *ptr != nil {
			fmt.Println("Overwriting entity at", inspector.Inspect(pos))
		}

		// should be safe as we've made sure the plot was large enough
		// to contain this entity
		*ptr = e

		return
	})
}

// Important note for understanding some of the math here:
// len(ring) = ringIdx * 8 + 4
//
// The greater value of |x| and |y| indicates which edge
// of the ring at which a coordinate is at.  Additionally,
// this value happens to be half the size of each edge -
// so we can always calculate corner positions from the length
// of the ring.
//
// distanceFromCenter = (len(ring) - 4) / 8 + .5
type Ring []Entity

func ensureNumber(e Entity) Entity {

	e.Number()

	return e
}

type BPNetwork struct {
	placementStarted bool
	connections      []BPConnection
	isGreen          bool
	placed           int
}

func (net *BPNetwork) createWire(conn1, conn2 BPConnection) {
	conn1.connector().addConnection(net, conn2)
	conn2.connector().addConnection(net, conn1)
}

type BPConnection struct {
	ent         ConnectorEntity
	isSecondary bool
}

func (b BPConnection) getCircuitId() int {
	if b.isSecondary {
		return 2
	}

	return 1
}

func (b BPConnection) connector() *EntityConnector {
	if b.isSecondary {
		return b.ent.(DualConnectorEntity).SecondaryConnector()
	}

	return b.ent.PrimaryConnector()
}

var distanceLookup10 = [...]int{
	10,
	10,
	9,
	9,
	9,
	9,
	8,
	7,
	6,
	5,
	1,
}

func iterateEdges(distance, step float64, cb func(pos Position)) {

	for x := -distance + step; x < distance; x += step {
		cb(Position{x, distance})
		cb(Position{x, -distance})
		cb(Position{distance, x})
		cb(Position{-distance, x})
	}

	cb(Position{distance, distance})
	cb(Position{distance, -distance})
	cb(Position{-distance, distance})
	cb(Position{-distance, -distance})
}

func (s *SubstationEntity) canConnectTo(pos Position) bool {
	if s.Position.X == pos.X {
		return math.Abs(s.Position.Y-pos.Y) <= 18
	}

	if s.Position.Y == pos.Y {
		return math.Abs(s.Position.X-pos.X) <= 18
	}

	return false
}

func (p Plot) getSubstation(neighborPos Position, from Position) (subst *SubstationEntity) {
	neighborPos.getBounds(1, 1).iterate(func(pos Position) (shouldBreak bool) {
		ent := p.get(pos)

		switch v := ent.(type) {
		case *SubstationEntity:
			if !v.canConnectTo(from) {
				return
			}

			subst = v
			shouldBreak = true
		case *RoboportEntity:
			switch {
			case neighborPos.X == from.X && from.Y > v.Pos().Y:
				subst, shouldBreak = p.get(Position{neighborPos.X, v.Pos().Y + 3}).(*SubstationEntity)
			case neighborPos.X == from.X && from.Y < v.Pos().Y:
				subst, shouldBreak = p.get(Position{neighborPos.X, v.Pos().Y - 3}).(*SubstationEntity)
			case neighborPos.Y == from.Y && from.X > v.Pos().X:
				subst, shouldBreak = p.get(Position{v.Pos().X + 3, neighborPos.Y}).(*SubstationEntity)
			case neighborPos.Y == from.Y && from.X < v.Pos().X:
				subst, shouldBreak = p.get(Position{v.Pos().X - 3, neighborPos.Y}).(*SubstationEntity)
			}
		}

		return

	})

	return
}

func (p *Plot) getAdjacentSubstation(pos Position) (stations []*SubstationEntity) {
	if subst := p.getSubstation(pos.shiftX(18), pos); subst != nil {
		stations = append(stations, subst)
	}

	if subst := p.getSubstation(pos.shiftX(-18), pos); subst != nil {
		stations = append(stations, subst)
	}

	if subst := p.getSubstation(pos.shiftY(18), pos); subst != nil {
		stations = append(stations, subst)
	}

	if subst := p.getSubstation(pos.shiftY(-18), pos); subst != nil {
		stations = append(stations, subst)
	}

	return

}

func (p *Plot) placeIrregularSubstation(at, original Position, wasPlacedVertical, isOffsetPositive bool) (subst *SubstationEntity) {
	subst = &SubstationEntity{}
	p.place(subst, at, DirectionDefault)

	switch {
	case wasPlacedVertical && isOffsetPositive:
		if neighbor := p.getSubstation(original.shiftY(18), original); neighbor != nil {
			subst.addNeighbor(neighbor)
		}
	case wasPlacedVertical && !isOffsetPositive:
		if neighbor := p.getSubstation(original.shiftY(-18), original); neighbor != nil {
			subst.addNeighbor(neighbor)
		}
	case !wasPlacedVertical && isOffsetPositive:
		if neighbor := p.getSubstation(original.shiftX(18), original); neighbor != nil {
			subst.addNeighbor(neighbor)
		}
	case !wasPlacedVertical && !isOffsetPositive:
		if neighbor := p.getSubstation(original.shiftX(-18), original); neighbor != nil {
			subst.addNeighbor(neighbor)
		}
	}

	return
}

func (p *Plot) expand() {
	var (
		oldLen = len(p.rings)
		newLen = len(p.rings) + 1

		substationDistance, roboportDistance float64
	)

	if newLen%18 == 0 {
		substationDistance = float64(newLen)
		newLen += 2
	}

	// new roboports required
	if newLen%50 >= 48 || newLen%50 == 0 {
		switch newLen % 50 {
		case 48:
			roboportDistance = float64(newLen + 2)
		case 49:
			roboportDistance = float64(newLen + 1)
		case 0:
			roboportDistance = float64(newLen)
		}

		for i := 0; i < 6; i++ {
			newLen++

			if substationDistance == 0 && newLen%18 == 0 {
				substationDistance = float64(newLen)
				newLen += 2
			}
		}

	}

	rings := make([]Ring, newLen-oldLen)

	for i := range rings {
		rings[i] = make(Ring, (oldLen+i)*8+4)
	}

	p.rings = append(p.rings, rings...)

	// . . .
	// SDWf

	if roboportDistance != 0 {
		iterateEdges(roboportDistance, 50, func(pos Position) {
			// In theory, we can halve roboport usage via this code.  Unfortunately, sometimes the substations which power
			// the roboports will be outside of the construction area of adjacent roboports - so the roboport won't be powered
			// and the blueprint might have issues with self-assembling.
			// if int(math.Abs(pos.X)) % 100 != int(math.Abs(pos.Y)) % 100 {
			// 	return
			// }

			p.place(ensureNumber(&RoboportEntity{}), pos, DirectionDefault)
		})
	}

	if substationDistance != 0 {
		iterateEdges(substationDistance, 18, func(pos Position) {
			bounds := pos.getBounds(1, 1)

			var roboport *RoboportEntity

			bounds.iterate(func(pos Position) bool {
				if ent := p.get(pos); ent != nil {
					var ok bool
					if roboport, ok = ent.(*RoboportEntity); !ok {
						panic(fmt.Errorf("expected roboport; got %s", inspector.Inspect(ent)))
					}

					return true
				}

				return false
			})

			if roboport != nil {
				roboPos := roboport.Pos()
				x, y := pos.abs()

				if y > x {
					subst := p.placeIrregularSubstation(Position{roboPos.X + 3, pos.Y}, pos, false, true)
					subst.addNeighbor(p.placeIrregularSubstation(Position{roboPos.X - 3, pos.Y}, pos, false, false))

					return
				}

				subst := p.placeIrregularSubstation(Position{pos.X, roboPos.Y + 3}, pos, true, true)
				subst.addNeighbor(p.placeIrregularSubstation(Position{pos.X, roboPos.Y - 3}, pos, true, false))

				return
			}

			subst := &SubstationEntity{}

			p.place(subst, pos, DirectionDefault)

			for _, neighbor := range p.getAdjacentSubstation(pos) {
				subst.addNeighbor(neighbor)
			}
		})
	}
}

func (p Plot) getRing(radius float64) Ring {
	r := int(radius)

	if r >= len(p.rings) {
		return nil
	}

	return p.rings[r]
}

// Since this math is a lil' bit complex it's illustrated here:
// https://www.desmos.com/calculator/jy9bj0dmsd
func (p Plot) getPtr(pos Position) *Entity {

	if pos.X == math.Floor(pos.X) {

		pos.X += .5
	}

	if pos.Y == math.Floor(pos.Y) {
		pos.Y += .5
	}

	area, radius := pos.getRingArea()

	ring := p.getRing(radius)

	if ring == nil {
		return nil
	}

	return ring.getPtr(area, radius, pos)
}

func (r Ring) getPtr(area RingArea, radius float64, pos Position) *Entity {
	switch area {
	case RingAreaBottomLeft:
		return &r[0]
	case RingAreaTopLeft:
		return &r[1]
	case RingAreaBottomRight:
		return &r[2]
	case RingAreaTopRight:
		return &r[3]
	case RingAreaLeft:
		return &r[int(radius+3+pos.Y)]
	case RingAreaRight:
		return &r[int(radius*3+2+pos.Y)]
	case RingAreaBottom:
		return &r[int(5*radius+pos.X+1)]
	case RingAreaTop:
		return &r[int(7*radius+pos.X)]
	}

	panic("invalid area")
}

func (r Ring) getCorner(pos Position) *Entity {
	if pos.X < 0 {
		if pos.Y < 0 {
			return &r[0]
		}

		return &r[1]
	}

	if pos.Y < 0 {
		return &r[2]
	}

	return &r[3]
}

func (r Ring) getVertical(pos Position, base float64) *Entity {
	if pos.X > 0 {
		return &r[int(base*3+2+pos.Y)]
	}

	return &r[int(base+3+pos.Y)]
}

func (r Ring) getHorizontal(pos Position, base float64) *Entity {
	if pos.Y > 0 {
		return &r[int(7*base+pos.X)]
	}

	return &r[int(5*base+pos.X+1)]
}

func (p Plot) get(pos Position) Entity {
	ptr := p.getPtr(pos)

	if ptr == nil {
		return nil
	}

	return *p.getPtr(pos)
}

func (p Plot) adjacent(pos Position) (e1, e2 Entity) {
	x, y := pos.abs()

	switch {
	case x == y:
		if pos.X < 0 {
			e1 = p.get(pos.shiftX(1))
		} else {
			e1 = p.get(pos.shiftX(-1))
		}

		if pos.Y < 0 {
			e2 = p.get(pos.shiftY(1))
		} else {
			e2 = p.get(pos.shiftY(-1))
		}
	case x > y:
		e1 = p.get(pos.shiftY(1))
		e2 = p.get(pos.shiftY(-1))
	case y > x:
		e1 = p.get(pos.shiftX(1))
		e2 = p.get(pos.shiftX(-1))
	}

	return
}

func (r Ring) sizeEven() (vertical, horizontal int) {
	base := (len(r) - 4) / 4

	return base + 2, base
}

func (r Ring) radius() float64 {
	return float64((len(r)-4)/8) + .5
}

func (r Ring) getNext(pl int) (p Position, d Direction, isFull bool) {
	v, h := r.sizeEven()

	radius := r.radius()

	defer func() {
		if p.toRingIdx()%2 == 0 {
			tmp := p.X
			p.X = p.Y
			p.Y = tmp

			d += 6

			d %= 8
		}

		// fmt.Println("rad", radius, "pl", pl, inspector.Inspect(p), inspector.Inspect(d))

	}()

	// int math is totally fine here (if pl isn't whole it will be floored - which is fine
	// since v is an int and any value which would be less after floored would be lesser anyways)
	if pl < v { // left
		yCoord := float64(pl) - radius + .5

		return Position{-radius, yCoord}, DirectionSouth, false
	}

	pl -= v

	if pl < h { // top
		xCoord := float64(pl) - radius + 1.5

		return Position{xCoord, radius}, DirectionEast, false
	}

	pl -= h

	if pl < v { // right
		// place top->bottom
		yCoord := -(float64(pl) - radius + .5)

		return Position{radius, yCoord}, DirectionNorth, false
	}

	pl -= v

	// bottom
	// place right->left
	xCoord := -(float64(pl) - radius + 1.5)

	return Position{xCoord, -radius}, DirectionWest, pl+2 == h
}

type EntityConnector struct {
	Red              []Connection `json:"red,omitempty"`
	Green            []Connection `json:"green,omitempty"`
	redNet, greenNet *BPNetwork
}

func (e *EntityConnector) undoConnection(net *BPNetwork) {
	if net.isGreen {
		e.Green = e.Green[:len(e.Green)-1]
		return
	}

	e.Red = e.Red[:len(e.Red)-1]
}

func (e *EntityConnector) addConnection(net *BPNetwork, b BPConnection) {
	conn := Connection{
		CircuitID: b.getCircuitId(),
		EntityID:  b.ent.Number(),
	}

	if net.isGreen {
		e.Green = append(e.Green, conn)
		return
	}

	e.Red = append(e.Red, conn)
	return
}

func (p Position) toRingIdx() int {
	x, y := p.abs()

	if x >= y {
		return int(x)
	}

	return int(y)
}

func (p Position) shift(x, y float64) Position {
	return Position{
		X: p.X + x,
		Y: p.Y + y,
	}
}

func (p Position) shiftX(d float64) Position {
	return Position{p.X + d, p.Y}
}

func (p Position) shiftY(d float64) Position {
	return Position{p.X, p.Y + d}
}

func (p Position) abs() (x, y float64) {
	return math.Abs(p.X), math.Abs(p.Y)
}

func (p Position) distanceXY(o Position) (x float64, y float64) {
	x = math.Abs(p.X - o.X)
	y = math.Abs(p.Y - o.Y)
	return
}

func (p Position) canConnect(o Position) bool {
	disX, disY := p.distanceXY(o)

	if int(disX) >= len(distanceLookup10) {
		return false
	}

	return int(disY) < distanceLookup10[int(disX)]
}

func (p Position) getBounds(length, width float64) Bounds {
	return Bounds{
		TR: Position{
			X: p.X + width,
			Y: p.Y + length,
		},
		BL: Position{
			X: p.X - width,
			Y: p.Y - length,
		},
	}
}

type Bounds struct {
	TR, BL Position
}

func (b Bounds) iterate(cb func(pos Position) bool) {
	var pos Position

	for pos.X = b.BL.X + .5; pos.X < b.TR.X; pos.X++ {
		for pos.Y = b.BL.Y + .5; pos.Y < b.TR.Y; pos.Y++ {
			if cb(pos) {
				break
			}
		}
	}
}

func (b Bounds) overlaps(o Bounds) bool {
	return !(b.TR.X < o.BL.X || b.TR.Y < o.BL.Y ||
		b.BL.X > o.TR.X || b.BL.Y > o.BL.Y)
}

func (b Bounds) contains(p Position) bool {
	return b.TR.X >= p.X && b.TR.Y >= p.Y &&
		b.BL.X <= p.X && b.BL.Y <= p.Y
}

type Connectors struct {
	// "Primary" connector - only connector for most components;
	// input for binary components.
	Primary *EntityConnector `json:"1,omitempty"`
	// "Secondary" connector - output connector for binary components.
	Secondary *EntityConnector `json:"2,omitempty"`
}

type Named interface {
	name() string
}

type Name[N Named] struct{}

func (n Name[N]) MarshalJSON() ([]byte, error) {
	var v N

	return json.Marshal(v.name())
}

type BaseEntity struct {
	EntityNumber int       `json:"entity_number"`
	Direction    Direction `json:"direction,omitempty"`
	Position     Position  `json:"position"`
}

func (b BaseEntity) wasPlaced() bool {
	return b.Position != Position{}
}

func (b BaseEntity) Bounds() Bounds {
	return b.Position.getBounds(b.getSize(b.Direction))
}

func (b BaseEntity) getSize(Direction) (l, w float64) {
	return .5, .5
}

type Direction int

type SingleConnectorBase struct {
	Connections *Connectors `json:"connections,omitempty"`
}

func (b *SingleConnectorBase) PrimaryConnector() *EntityConnector {
	if b.Connections == nil {
		b.Connections = new(Connectors)
	}

	if b.Connections.Primary == nil {
		b.Connections.Primary = new(EntityConnector)
	}

	return b.Connections.Primary
}

type DualConnectorsBase struct{ SingleConnectorBase }

func (b *DualConnectorsBase) SecondaryConnector() *EntityConnector {
	if b.Connections == nil {
		b.Connections = new(Connectors)
	}

	if b.Connections.Secondary == nil {
		b.Connections.Secondary = new(EntityConnector)
	}

	return b.Connections.Secondary
}

func (d Direction) isVertical() bool {
	return d == DirectionNorth || d == DirectionSouth
}

func (d Direction) InspectCustom() inspector.InspectString {
	switch d {
	case DirectionNorth:
		return "N"
	case DirectionEast:
		return "E"
	case DirectionSouth:
		return "S"
	case DirectionWest:
		return "W"
	default:
		return "invalid direction"
	}
}

const (
	DirectionNorth Direction = iota * 2
	DirectionEast
	DirectionSouth
	DirectionWest

	// DirectionVertical   Direction = iota * 2 // north
	// DirectionHorizontal                      // east

	DirectionDefault = DirectionNorth
)

var entityNumber = 0

func (e *BaseEntity) basic() *BaseEntity {
	return e
}

func (e *BaseEntity) wireLength() float64 {
	return 10
}

func (e *BaseEntity) Pos() Position {
	return e.Position
}

func (e *BaseEntity) Number() int {
	if e.EntityNumber == 0 {
		entityNumber++
		e.EntityNumber = entityNumber
	}

	return e.EntityNumber
}

type Entity interface {
	Named
	Number() int
	Bounds() Bounds
	getSize(d Direction) (length, width float64)
	basic() *BaseEntity
	wasPlaced() bool
	Pos() Position
}

type ConnectorEntity interface {
	Entity
	PrimaryConnector() *EntityConnector
}

type DualConnectorEntity interface {
	ConnectorEntity
	SecondaryConnector() *EntityConnector
}

type PowerEntity interface {
	PowerArea() Bounds
}

type LampEntity struct {
	BaseEntity
	Name            Name[LampEntity]    `json:"name"`
	ControlBehavior LampControlBehavior `json:"control_behavior"`
}

func (LampEntity) name() string {
	return "lamp"
}

type ArithmeticCombinatorEntity struct {
	BaseEntity
	DualConnectorsBase
	Name            Name[ArithmeticCombinatorEntity] `json:"name"`
	ControlBehavior ArithmeticControlBehavior        `json:"control_behavior"`
}

func (ArithmeticCombinatorEntity) name() string {
	return "arithmetic-combinator"
}

func (a ArithmeticCombinatorEntity) getSize(d Direction) (l, w float64) {
	if d.isVertical() {
		return 1, .5
	}

	return .5, 1
}

func (a ArithmeticCombinatorEntity) Bounds() Bounds {
	return a.Position.getBounds(a.getSize(a.Direction))
}

type DeciderCombinatorEntity struct {
	BaseEntity
	DualConnectorsBase
	Name            Name[DeciderCombinatorEntity] `json:"name"`
	ControlBehavior DeciderControlBehavior        `json:"control_behavior"`
}

func (DeciderCombinatorEntity) name() string {
	return "decider-combinator"
}

func (d DeciderCombinatorEntity) getSize(dir Direction) (l, w float64) {
	if dir.isVertical() {
		return 1, .5
	}

	return .5, 1
}

func (d DeciderCombinatorEntity) Bounds() Bounds {
	return d.Position.getBounds(d.getSize(d.Direction))
}

type ConstantCombinatorEntity struct {
	BaseEntity
	SingleConnectorBase
	Name            Name[ConstantCombinatorEntity] `json:"name"`
	ControlBehavior ConstantControlBehavior        `json:"control_behavior"`
}

func (ConstantCombinatorEntity) name() string {
	return "constant-combinator"
}

type SubstationEntity struct {
	BaseEntity
	Name      Name[SubstationEntity] `json:"name"`
	Neighbors []int                  `json:"neighbours,omitempty"`
}

func (SubstationEntity) name() string {
	return "substation"
}

func (s *SubstationEntity) addNeighbor(o *SubstationEntity) {
	s.Neighbors = append(s.Neighbors, o.Number())
	o.Neighbors = append(o.Neighbors, s.Number())
}

func (s *SubstationEntity) wasPlaced() bool {
	return true
}

func (s *SubstationEntity) getSize(Direction) (l, w float64) {
	return 1, 1
}

func (s *SubstationEntity) Bounds() Bounds {
	return s.Position.getBounds(s.getSize(s.Direction))
}

func (s *SubstationEntity) PowerArea() Bounds {
	return s.Position.getBounds(9, 9)
}

type Planner struct {
	Plot
	networks     []BPNetwork
	powerSources []PowerEntity

	// free []Position
	placed  int
	ringIdx int
	free    []Free2x1Location
}

type RoboportEntity struct {
	BaseEntity
	Name Name[RoboportEntity] `json:"name"`
}

func (RoboportEntity) name() string {
	return "roboport"
}

func (r RoboportEntity) getSize(Direction) (l, w float64) {
	return 2, 2
}

func (r *RoboportEntity) Bounds() Bounds {
	return r.Position.getBounds(r.getSize(DirectionDefault))
}

func (p *Planner) getNext2x1() (Position, Direction) {
	// offset for 1x1 placements
	if p.placed%2 == 1 {
		p.placed++
	}
	// b := e.basic()

	for {
		pos, dir, mustExpand := p.Plot.rings[p.ringIdx].getNext(p.placed)

		if mustExpand {
			p.ringIdx++
			p.placed = 0

			if p.ringIdx+1 == len(p.rings) {
				p.expand()
			}
		} else {
			p.placed += 2
		}

		var bounds Bounds

		if dir.isVertical() {
			bounds = pos.getBounds(1, .5)
		} else {
			bounds = pos.getBounds(.5, 1)
		}

		var freePos Position
		canPlace := true

		bounds.iterate(func(pos Position) bool {
			if p.get(pos) == nil {
				freePos = pos

				return false
			}

			canPlace = false

			return false
		})

		if canPlace {
			return pos, dir
		}

		if freePos.isZero() {
			continue
		}

		area, _ := pos.getRingArea()

		switch area {
		case RingAreaLeft:
			newPos := pos.shift(-.5, freePos.Y-pos.Y)
			dir = DirectionEast

			if p.canPlace(newPos.getBounds(.5, 1)) {
				return newPos, dir
			}
		case RingAreaRight:
			newPos := pos.shift(-.5, freePos.Y-pos.Y)
			dir = DirectionWest

			if p.canPlace(newPos.getBounds(.5, 1)) {
				return newPos, dir
			}
		case RingAreaBottom:
			newPos := pos.shift(freePos.X-pos.X, .5)
			dir = DirectionNorth

			if p.canPlace(newPos.getBounds(.5, 1)) {
				return newPos, dir
			}
		case RingAreaTop:
			newPos := pos.shift(freePos.X-pos.X, -.5)
			dir = DirectionSouth

			if p.canPlace(newPos.getBounds(.5, 1)) {
				return newPos, dir
			}
		}

	}
}

func (p *Planner) getNext1x1() Position {
	for {
		pos, _, mustExpand := p.Plot.rings[p.ringIdx].getNext(p.placed)

		if mustExpand {
			p.ringIdx++
			p.placed = 0

			if p.ringIdx+1 == len(p.rings) {
				p.expand()
			}
		} else {
			p.placed++
		}

		if p.get(pos) == nil {
			return pos
		}
	}
}

func (p *Planner) getNext(e Entity) (Position, Direction) {
	l, w := e.getSize(DirectionNorth)

	if l == .5 && w == .5 {
		return p.getNext1x1(), DirectionNorth
	}

	if l == 1 && w == .5 {
		return p.getNext2x1()
	}

	panic("unable to get next")
}

func (p *Planner) tryConnect(entConn BPConnection, net *BPNetwork) bool {
	// for i := len(net.connections) - 1; i >= 0; i-- {
	// 	conn := net.connections[i]
	for _, conn := range net.connections {
		if !conn.ent.wasPlaced() || conn == entConn {
			continue
		}

		if conn.ent.Pos().canConnect(entConn.ent.Pos()) {
			net.createWire(entConn, conn)

			return true
		}
	}

	return false
}

type Free2x1Location struct {
	Position  Position
	Direction Direction
	Attempts  int
}

// func (p *Planner) tryPlaceAndConnect(e Entity, n *BPNetwork) {

// }

// func (p *Planner) tryReplace(target Entity, with Entity) bool {

// }

// future ideas for placement:
// 1. when network is started, we'll store the area
// where future points must be place-able in.  Each time we place we check to see which networks can be placed
// in said area and try to place the one with least options.  In the event that 2 only have 1 option, we should re-arrange
// to accommodate both

func (p *Planner) placeNetwork(net *BPNetwork) {
	net.placementStarted = true

	// var placedIdx Entity

	for i, conn := range net.connections {
		var attempts int
		ent := conn.ent
		if ent.Number() == debugEntityId {
			fmt.Println("hm", i)
			fmt.Println(ent.PrimaryConnector())
		}

		b := ent.basic()
		wasPlaced := ent.wasPlaced()

	tryAgain:
		if !wasPlaced {
			b.Position, b.Direction = p.getNext2x1()
			p.place(ent, b.Position, b.Direction)
		}

		if i == 0 {
			continue
		}

		if !p.tryConnect(conn, net) {
			// fmt.Println(inspector.Inspect(conn))
			// TODO: be smart about this + retrying
			fmt.Println(ent.Pos(), net.connections[0].ent.Pos())

			if attempts < 2000 {
				attempts++
				goto tryAgain
			}

			panic("unable to connect")
		}

		if attempts > 0 {
			fmt.Println("placing on attempt #", attempts)
		}

		// lastPlaced := net.connection

		// if n := ent.PrimaryConnector().

		// p.Plot[p.plotIdx]
		// 1. get next free position
		// 2. check for adjacent entities
		// 3. place if available, otherwise expand and goto 1.

		// b := conn.Entity.basic()

		// if p.placed <
	}

	// if (p) {}
	//
	// if len(net.connections) == 1 {

	// }

}

func (p *Planner) Save(to io.Writer) {
	to.Write([]byte{'0'})
	enc := base64.NewEncoder(base64.StdEncoding, to)
	defer enc.Close()
	comp := zlib.NewWriter(enc)
	defer comp.Close()

	out := json.NewEncoder(comp)

	// ensure outer rings are powered
	for len(p.rings)%18 > 9 {
		p.expand()
	}
	inspector.Println(p)

	out.Encode(Blueprint{BlueprintData{
		Icons:    []BlueprintIcon{{Signal: &Signal{SignalTypeVirtual, "signal-everything"}, Index: 1}},
		Entities: p.entities,
		Item:     "blueprint",
		Version:  281479274168320,
	}})
}

func (p *Planner) DoLargeBuild() {
	p.Plot = Plot{
		rings: []Ring{
			make(Ring, 4),
			make(Ring, 4+1*8),
			make(Ring, 4+2*8),
			make(Ring, 4+3*8),
			make(Ring, 4+4*8),
		},
	}
	p.ringIdx = 2

	p.place(&RoboportEntity{}, Position{}, DirectionDefault)

	p.placeIrregularSubstation(Position{0, 3}, Position{}, true, true).
		addNeighbor(p.placeIrregularSubstation(Position{0, -3}, Position{}, true, false))

	var net *BPNetwork

	// for i := 0; i < 100; i++ {
	// 	p.expand()
	// }

	for i := 0; i < 2000; i++ {
		var e DualConnectorEntity

		if i%2 == 0 {
			e = &ArithmeticCombinatorEntity{ControlBehavior: ArithmeticControlBehavior{ArithmeticConditions{
				ArithmeticInputs: ArithmeticInputs{
					FirstSignal: signals[0],
					Constant:    1,
				},
				Operation:    ArithmeticOperationAdd,
				OutputSignal: signals[0],
			}}}
		} else {
			e = &DeciderCombinatorEntity{ControlBehavior: DeciderControlBehavior{DeciderConditions{
				BooleanCondition: BooleanCondition{
					FirstSignal: signals[0],
					Constant:    0,
					Comparator:  ComparatorEq,
				},
				OutputSignal:       signals[0],
				CopyCountFromInput: true,
			}}}
		}

		if net != nil {
			net.connections[1] = BPConnection{ent: e}
		}

		pos, dir := p.getNext2x1()
		p.place(e, pos, dir)
	}

	// Position{0, 0}.getBounds(1, 1).iterate(func(pos Position) bool {
	// 	inspector.Println(pos)
	// 	return false
	// })

	return
}

// func (p *Planner) SanityCheck() {
// 	p.Plot = Plot{
// 		rings: []Ring{
// 			make(Ring, 4),
// 			make(Ring, 4+1*8),
// 			make(Ring, 4+2*8),
// 			make(Ring, 4+3*8),
// 			make(Ring, 4+4*8),
// 		},
// 	}

// }

func DoLargeBuild(o io.Writer) {
	const bits = 12
	const layers = bits - 1

	var nets = make([]*Network, bits)

	sig := signals[1]
	ic := InputComponent{
		primaryInput:   sig,
		secondaryInput: Constant(0),
	}

	d := &Decider{
		InputComponent: ic,
		Operator:       ComparatorEq,
		output:         sig,

		outputFixed: true,
	}

	nets[0].connectOutput(d)

	for i := 0; i < layers; i++ {
		in := nets[i]
		out := nets[i+1]

		cnt := 1 << (layers - i)

		for n := 0; n < cnt; n++ {
			d := &Decider{
				InputComponent: ic,
				Operator:       ComparatorNe,
				output:         sig,
				outputFixed:    true,
			}

			in.connectInput(d)

			out.connectOutput(d)
		}

		d := &Decider{

			InputComponent: ic,
			Operator:       ComparatorNe,
			output:         sig,
		}

		in.connectInput(d)
		out.connectOutput(d)
	}

	end := &Arithmetic{
		InputComponent: InputComponent{
			primaryInput:   sig,
			secondaryInput: Constant(1),
		},
		Operator: ArithmeticOperationAdd,
		output:   sig,
	}
	nets[layers].connectInput(end)

	pl := &Planner{}

	pl.plan(nets)

	pl.Save(o)
}

type placerStackItem struct {
	*stackItem
	p         Position
	d         Direction
	idx       int
	undoFuncs []func(item *placerStackItem)
	attempted map[Entity]struct{}
	sids      []int
}

func (p *placerStackItem) addUndo(fn func(item *placerStackItem)) {
	p.undoFuncs = append(p.undoFuncs, fn)
}

func (p *placerStackItem) undo() {
	for i := len(p.undoFuncs) - 1; i >= 0; i-- {
		p.undoFuncs[i](p)
	}

	p.undoFuncs = nil
}

func (p *placerStackItem) connect(e ConnectorEntity, net *BPNetwork, isSecondary bool) bool {
	newConn := BPConnection{
		ent:         e,
		isSecondary: isSecondary,
	}

	net.connections = append(net.connections, newConn)
	p.addUndo(func(*placerStackItem) {
		net.connections = net.connections[:len(net.connections)-1]
	})

	if len(net.connections) == 1 {
		return true
	}

	for _, conn := range net.connections[:len(net.connections)-1] {
		if e.Pos().canConnect(conn.ent.Pos()) {
			net.createWire(newConn, conn)
			p.addUndo(func(*placerStackItem) {
				newConn.connector().undoConnection(net)
				conn.connector().undoConnection(net)
			})
			return true
		}
	}

	return false
}

func (p *placerStackItem) makeConnections(e ConnectorEntity) bool {
	conn := e.PrimaryConnector()

	if conn.redNet != nil {
		if !p.connect(e, conn.redNet, false) {
			return false
		}
	}

	if conn.greenNet != nil {
		if !p.connect(e, conn.greenNet, false) {
			return false
		}
	}

	if dce, ok := e.(DualConnectorEntity); ok {
		conn = dce.SecondaryConnector()

		if conn.redNet != nil {
			if !p.connect(e, conn.redNet, true) {
				return false
			}
		}

		if conn.greenNet != nil {
			if !p.connect(e, conn.greenNet, true) {
				return false
			}
		}
	}

	return true
}

func (p *Planner) planSlow(entities []Entity) {
	var (
		entStackHead = stackItem{}
		tail         = &entStackHead
		placed       = map[Entity]struct{}{}
		stack        = []placerStackItem{}
		popped       = []placerStackItem{}
	)

	for _, ent := range entities {
		tail = tail.add(ent)
	}

	var item placerStackItem

	item.p, item.d = p.getNext2x1()
	item.stackItem = &entStackHead
	item.attempted = make(map[Entity]struct{})

stackHead:
	didPlace := false
	item.stackItem.iterate(func(s *stackItem) (shouldBreak bool) {

		item.undo()
		item.sids = append(item.sids, s.id)
		item.attempted[s.ent] = struct{}{}

		if _, wasPlaced := placed[item.ent]; wasPlaced {
			return
		}

		item.stackItem = s

		placed[item.ent] = struct{}{}
		item.stackItem.remove()

		item.addUndo(func(item *placerStackItem) {
			delete(placed, item.ent)
			item.reinstate()
		})

		l, w := item.ent.getSize(item.d)
		var offX, offY float64

		if l == .5 && w == .5 {
			if item.d.isVertical() {
				offY = .5
			} else {
				offX = .5
			}
		}

		p.place(item.ent, item.p.shift(offX, offY), item.d)
		item.addUndo(func(item *placerStackItem) {
			p.remove(item.ent)
		})

		if c, ok := item.ent.(ConnectorEntity); ok {
			// if item.p.X == 5.5 && item.p.Y == -5 {
			// 	fmt.Println(c.PrimaryConnector())
			// }
			if !item.makeConnections(c) {
				return false
			}
		}

		stack = append(stack, item)

		if len(popped) > 0 {
			item = popped[len(popped)-1]
			popped = popped[:len(popped)-1]
		} else {
			item = placerStackItem{}
			item.p, item.d = p.getNext2x1()
		}

		item.stackItem = &entStackHead
		item.attempted = make(map[Entity]struct{})
		shouldBreak = true
		didPlace = true
		return
	})
	// fmt.Println("placed", len(placed), len(entities), stack[len])
	if didPlace && entStackHead.next != nil {
		goto stackHead
	}

	if entStackHead.next != nil || !didPlace {
		item.undo()
		popped = append(popped, item)
		item = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		goto stackHead
	}
}

func (p *Planner) plan(networks []*Network) {
	p.Plot = Plot{
		rings: []Ring{
			make(Ring, 4),
			make(Ring, 4+1*8),
			make(Ring, 4+2*8),
			make(Ring, 4+3*8),
			make(Ring, 4+4*8),
		},
	}

	p.ringIdx = 2

	p.place(&RoboportEntity{}, Position{}, DirectionDefault)

	p.placeIrregularSubstation(Position{0, 3}, Position{}, true, true).
		addNeighbor(p.placeIrregularSubstation(Position{0, -3}, Position{}, true, false))

	_ = networks[0]
	p.networks = make([]BPNetwork, len(networks))

	// p.place(&RoboportEntity{}, Position{0, 0}, DirectionDefault)
	//
	entm := map[Entity]struct{}{}
	ents := []Entity{}

	for i, net := range networks {
		bpnet := &p.networks[i]

		// bpnet.connections = make([]BPConnection, len(net.members))
		bpnet.isGreen = net.isGreen

		for _, c := range net.members {
			ent := c.Component.Entity()
			conn := BPConnection{
				ent:         ent,
				isSecondary: c.isSecondary,
			}

			if bpnet.isGreen {
				conn.connector().greenNet = bpnet
			} else {
				conn.connector().redNet = bpnet
			}

			if _, ok := entm[ent]; !ok {
				entm[ent] = struct{}{}
				ents = append(ents, ent)
			}

		}
	}

	p.planSlow(ents)
}
