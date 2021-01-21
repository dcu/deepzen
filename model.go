package tabnet

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gonum.org/v1/plot/vg"
	"gorgonia.org/gorgonia"
	"gorgonia.org/gorgonia/encoding/dot"
	"gorgonia.org/qol/plot"
	"gorgonia.org/tensor"
)

const (
	heatmapPath     = "heatmap"
	bufferSizeModel = 16
)

// TrainOpts are the options to train the model
type TrainOpts struct {
	Epochs    int
	BatchSize int

	DevMode               bool
	WithLearnablesHeatmap bool

	// Solver defines the solver to use. It uses gorgonia.AdamSolver by default if none is passed
	Solver gorgonia.Solver

	CostFn func(output *gorgonia.Node, loss *gorgonia.Node, y *gorgonia.Node) *gorgonia.Node
}

func (o *TrainOpts) setDefaults() {
	if o.Epochs == 0 {
		o.Epochs = 10
	}

	if o.BatchSize == 0 {
		o.BatchSize = 1024
	}
}

// Model implements the tab net model
type Model struct {
	g          *gorgonia.ExprGraph
	learnables gorgonia.Nodes
	watchables map[string]*gorgonia.Value

	model map[string]gorgonia.Value
}

// NewModel creates a new model for the neural network
func NewModel() *Model {
	return &Model{
		g:          gorgonia.NewGraph(),
		learnables: make([]*gorgonia.Node, 0, bufferSizeModel),
		watchables: make(map[string]*gorgonia.Value),
		model:      make(map[string]gorgonia.Value, bufferSizeModel),
	}
}

// ToSVG creates a SVG representation of the node
func (m *Model) ToSVG(path string) error {
	b, err := dot.Marshal(m.g)
	if err != nil {
		return err
	}

	fileName := "graph.dot"

	err = ioutil.WriteFile(fileName, b, 0644)
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(fileName) }()

	cmd := exec.Command("dot", "-T", "svg", fileName, "-o", path)

	return cmd.Run()
}

func (m *Model) ExprGraph() *gorgonia.ExprGraph {
	return m.g
}

func (m *Model) Train(layer Layer, trainX tensor.Tensor, trainY tensor.Tensor, opts TrainOpts) error {
	opts.setDefaults()

	if opts.DevMode {
		warn("Start training in dev mode")
	}

	if opts.WithLearnablesHeatmap {
		if opts.BatchSize > 128 {
			panic("to enable Heatmap BatchSize must be <= 128")
		}

		warn("Heatmaps will be stored in: %s", heatmapPath)
		_ = os.RemoveAll(heatmapPath)
	}

	numExamples, features := trainX.Shape()[0], trainX.Shape()[1]
	batches := numExamples / opts.BatchSize

	x := gorgonia.NewTensor(m.g, tensor.Float64, trainX.Shape().Dims(), gorgonia.WithShape(opts.BatchSize, features), gorgonia.WithName("x"))
	y := gorgonia.NewMatrix(m.g, tensor.Float64, gorgonia.WithShape(opts.BatchSize, trainY.Shape()[1]), gorgonia.WithName("y"))

	output, loss, err := layer(x)
	if err != nil {
		return fmt.Errorf("error running layer: %w", err)
	}

	if loss == nil {
		return fmt.Errorf("loss must be returned in training mode")
	}

	var (
		costVal gorgonia.Value
		predVal gorgonia.Value
	)

	{
		cost := opts.CostFn(output, loss, y)

		gorgonia.Read(cost, &costVal)
		gorgonia.Read(output, &predVal)

		if _, err := gorgonia.Grad(cost, m.learnables...); err != nil {
			return fmt.Errorf("error calculating gradient: %w", err)
		}
	}

	vmOpts := []gorgonia.VMOpt{
		gorgonia.BindDualValues(m.learnables...),
	}

	if opts.DevMode {
		vmOpts = append(
			vmOpts,
			gorgonia.TraceExec(),
			gorgonia.WithNaNWatch(),
			gorgonia.WithInfWatch(),
		)
	}

	vm := gorgonia.NewTapeMachine(m.g, vmOpts...)

	if opts.Solver == nil {
		// opts.Solver = gorgonia.NewRMSPropSolver(gorgonia.WithBatchSize(float64(opts.BatchSize)))
		opts.Solver = gorgonia.NewAdamSolver(gorgonia.WithBatchSize(float64(opts.BatchSize)), gorgonia.WithLearnRate(0.02))
	}

	defer vm.Close()

	// bar := pb.New(batches)
	// bar.SetRefreshRate(time.Second)
	// bar.SetMaxWidth(80)

	startTime := time.Now()

	for i := 0; i < opts.Epochs; i++ {
		// bar.Prefix(fmt.Sprintf("Epoch %d", i))
		// bar.Set(0)
		// bar.Start()

		for b := 0; b < batches; b++ {
			start := b * opts.BatchSize
			end := start + opts.BatchSize

			if start >= numExamples {
				break
			}

			if end > numExamples {
				end = numExamples
			}

			xVal, err := trainX.Slice(gorgonia.S(start, end))
			if err != nil {
				return err
			}

			yVal, err := trainY.Slice(gorgonia.S(start, end))
			if err != nil {
				return err
			}

			err = xVal.(*tensor.Dense).Reshape(opts.BatchSize, features)
			if err != nil {
				return err
			}

			err = gorgonia.Let(x, xVal)
			if err != nil {
				fatal("error assigning x: %v", err)
			}

			err = gorgonia.Let(y, yVal)
			if err != nil {
				fatal("error assigning y: %v", err)
			}

			if err = vm.RunAll(); err != nil {
				fatal("Failed at epoch  %d, batch %d. Error: %v", i, b, err)
			}

			if err = opts.Solver.Step(gorgonia.NodesToValueGrads(m.learnables)); err != nil {
				fatal("Failed to update nodes with gradients at epoch %d, batch %d. Error %v", i, b, err)
			}

			// color.Yellow(" Epoch %d %d | cost %v\n", i, b, costVal)

			m.PrintWatchables()

			if opts.WithLearnablesHeatmap {
				m.saveHeatmaps(i, b, opts.BatchSize, features)
			}

			vm.Reset()
			// bar.Increment()
		}

		fmt.Printf(" Epoch %d | cost %v (%v)\n", i, costVal, time.Since(startTime))
	}

	fmt.Println("")

	return nil
}

func (m Model) saveHeatmaps(epoch int, batch int, batchSize, features int) {
	for _, v := range m.learnables {
		wt := v.Value().(tensor.Tensor)
		wtShape := wt.Shape().Clone()
		newShape := tensor.Shape{wtShape[0], tensor.Shape(wtShape[1:]).TotalSize()}

		pathName := filepath.Join(heatmapPath, v.Name())
		fileName := fmt.Sprintf("%s/%d_%d_%v.png", pathName, epoch, batch, wtShape)

		err := wt.Reshape(newShape...)
		if err != nil {
			panic(err)
		}

		p, err := plot.Heatmap(wt, nil)
		if err != nil {
			panic(fmt.Errorf("failed to process %s: %w", fileName, err))
		}

		err = wt.Reshape(wtShape...)
		if err != nil {
			panic(err)
		}

		width := vg.Length(features) * vg.Centimeter
		height := vg.Length(batchSize) * vg.Centimeter

		_ = os.MkdirAll(pathName, 0755)
		_ = p.Save(width, height, fileName)
	}
}

func (m Model) checkArity(contextName string, nodes []*gorgonia.Node, arity int) error {
	if len(nodes) != arity {
		return fmt.Errorf("arity doesn't match on %s, expected %d, got %d", contextName, arity, len(nodes))
	}

	return nil
}
