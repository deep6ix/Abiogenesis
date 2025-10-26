package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	StepsPerTick = 100 // Speed up the simulation dramatically
)

// --- SIMULATION CORE (Pond, Molecule, Reaction remain largely the same) ---

// A simplified Molecule struct.
type Molecule struct {
	Name string
}

// A Reaction defines how molecules interact.
// If Catalyst is empty, it's a non-catalytic reaction.
// If Product equals Catalyst, it has the potential to be autocatalytic.
type Reaction struct {
	Reactants []string
	Product   string
	Catalyst  string
}

// Pond represents the state of the simulation environment.
type Pond struct {
	Molecules    map[string]int // Molecule Name -> Count
	Reactions    []Reaction
	LastReaction string // To display in the UI
}

// NewPond initializes the simulation with basic molecules and core reactions.
func NewPond() *Pond {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Define initial basic molecules and their counts (A, B, C are the 'food' molecules)
	initialMolecules := map[string]int{
		"A": 500, // Increased starting materials for faster CAS emergence
		"B": 500,
		"C": 500,
		"D": 0, // Complex molecule D (precursor)
		"E": 1, // Start with one 'E' to kick off the autocatalysis immediately
	}

	// Define core reactions.
	// 1. Basic formation (A + B -> D)
	// 2. CAS Initialization (D + C -> E) - Requires D and C to be present.
	// 3. Autocatalysis (D + A -> E, catalyzed by E) - The key self-reproducing reaction.
	// 4. Degradation (E -> C + B) - To prevent infinite growth.
	coreReactions := []Reaction{
		{Reactants: []string{"A", "B"}, Product: "D", Catalyst: ""},  // R1: Basic synthesis
		{Reactants: []string{"D", "C"}, Product: "E", Catalyst: ""},  // R2: Initial complex formation
		{Reactants: []string{"D", "A"}, Product: "E", Catalyst: "E"}, // R3: Autocatalysis
		{Reactants: []string{"E"}, Product: "A", Catalyst: ""},       // R4: Degradation/Recycling
	}

	return &Pond{
		Molecules:    initialMolecules,
		Reactions:    coreReactions,
		LastReaction: "Simulation Initialized",
	}
}

// Step runs one tick of the simulation.
func (p *Pond) Step() {
	if len(p.Reactions) == 0 {
		p.LastReaction = "No reactions defined."
		return
	}

	// 1. Select a random reaction to attempt
	r := p.Reactions[rand.Intn(len(p.Reactions))]

	// 2. Check reactants availability
	canReact := true
	for _, reactant := range r.Reactants {
		if p.Molecules[reactant] <= 0 {
			canReact = false
			break
		}
	}

	// 3. Check catalyst requirement
	if canReact && r.Catalyst != "" {
		// For catalyzed reactions, the catalyst must be present
		if p.Molecules[r.Catalyst] <= 0 {
			canReact = false
		}
	}

	// 4. Execute the reaction if possible
	if canReact {
		// Consume reactants
		for _, reactant := range r.Reactants {
			p.Molecules[reactant]--
		}

		// In this simplified model, we don't consume the catalyst.
		// If the catalyst is the product (Autocatalysis, R3), it's conserved.

		// Produce product
		p.Molecules[r.Product]++

		// Track reaction for UI
		reactantsStr := ""
		for i, rName := range r.Reactants {
			reactantsStr += rName
			if i < len(r.Reactants)-1 {
				reactantsStr += " + "
			}
		}

		catalystStr := ""
		if r.Catalyst != "" {
			catalystStr = fmt.Sprintf(" (Cat: %s)", r.Catalyst)
		}
		p.LastReaction = fmt.Sprintf("Reaction: %s -> %s%s", reactantsStr, r.Product, catalystStr)
	} else {
		// If a reaction fails, we keep the last successful event for better visualization clarity.
		// To avoid overwhelming the status display with constant "failed" messages, we skip the update.
	}
}

// --- Ebitengine Game Implementation ---

// Game implements ebiten.Game and holds the simulation state.
type Game struct {
	Pond        *Pond
	TickCounter int
}

func NewGame() *Game {
	return &Game{
		Pond: NewPond(),
	}
}

// Update updates the game state. This is where the simulation steps run.
func (g *Game) Update() error {
	// Run multiple simulation steps per frame for fast evolution
	for i := 0; i < StepsPerTick; i++ {
		g.Pond.Step()
	}
	g.TickCounter++
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black) // Dark background for contrast

	// Title
	title := "Autocatalytic Pond Simulation (Ebitengine)"
	text.Draw(screen, title, basicfont.Face7x13, 20, 30, color.White)

	// Simulation Status
	status := fmt.Sprintf("Sim Ticks: %d | Steps/Tick: %d", g.TickCounter, StepsPerTick)
	text.Draw(screen, status, basicfont.Face7x13, 20, 50, color.White)

	text.Draw(screen, "Last Event:", basicfont.Face7x13, 20, 70, color.RGBA{180, 180, 180, 255})
	text.Draw(screen, g.Pond.LastReaction, basicfont.Face7x13, 100, 70, color.White)

	// Molecule Visualization
	yOffset := 100
	xName := 20
	xCount := 120

	text.Draw(screen, "Molecule", basicfont.Face7x13, xName, yOffset, color.RGBA{100, 200, 255, 255})
	text.Draw(screen, "Count", basicfont.Face7x13, xCount, yOffset, color.RGBA{100, 200, 255, 255})

	yOffset += 20

	// Draw molecule counts, highlighting the critical CAS molecule 'E'
	for name, count := range g.Pond.Molecules {
		yOffset += 20

		// Color logic:
		molColor := color.White // Default for basic molecules (A, B, C)

		// Simple visual feedback: size of the rectangle represents molecule count
		rectMax := ScreenWidth - xCount - 150
		rectHeight := 15
		rectWidth := count / 5
		if rectWidth > rectMax {
			rectWidth = rectMax // Cap the bar width
		}
		if rectWidth < 0 {
			rectWidth = 0
		}

		barColor := color.RGBA{50, 50, 50, 150} // Default grey bar

		if name == "D" {
			molColor = color.RGBA{255, 255, 0, 255} // Yellow for Precursor
			barColor = color.RGBA{255, 255, 0, 100}
		} else if name == "E" {
			// Red/Orange highlight for the Autocatalytic product
			molColor = color.RGBA{255, 100, 50, 255}
			barColor = color.RGBA{255, 0, 0, 100} // Faded red bar

			// Check for CAS Emergence based on absolute count
			if count > 5000 {
				molColor = color.RGBA{0, 255, 0, 255} // Green when dominant
			}
		}

		// Draw the dynamic bar
		ebiten.DrawRect(screen, rectWidth, rectHeight, barColor, &ebiten.DrawRectOptions{
			GeoM: ebiten.Translate(float64(xCount+80), float64(yOffset-11)),
		})

		// Draw molecule name and count
		text.Draw(screen, name, basicfont.Face7x13, xName, yOffset, molColor)
		text.Draw(screen, strconv.Itoa(count), basicfont.Face7x13, xCount, yOffset, molColor)
	}

	// Final Emergence Message
	if g.Pond.Molecules["E"] > 5000 {
		emergenceText := fmt.Sprintf("!!! CAS DOMINANCE ACHIEVED (E: %d) !!!", g.Pond.Molecules["E"])
		text.Draw(screen, emergenceText, basicfont.Face7x13, xName, ScreenHeight-30, color.RGBA{0, 255, 0, 255})
	}
}

// Layout returns the screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// The new main function runs the Ebitengine game loop.
func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Go Autocatalytic Set - Ebitengine")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
