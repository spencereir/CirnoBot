package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var (
	S                    = []float64{1.0, 0.0, 0.0}
	A                    = []float64{0.0, 1.0, 0.0}
	B                    = []float64{0.0, 0.0, 1.0}
	NUM_HIDDEN_LAYERS    = 2
	HIDDEN_LAYER_SIZE    = 50
	INPUT_LAYER_SIZE     = 1200
	OUTPUT_LAYER_SIZE    = 3
	CONVERGENCE_ATTEMPTS = 500
	LEARNING_RATE        = 0.0001
	LAMBDA               = 2.0
	w                    = make([][][]float64, 1+NUM_HIDDEN_LAYERS)
	new_w                = make([][][]float64, 1+NUM_HIDDEN_LAYERS)
	a                    = make([][]float64, NUM_HIDDEN_LAYERS)
	delta                = make([][]float64, 2+NUM_HIDDEN_LAYERS)
	Delta                = make([][][]float64, 1+NUM_HIDDEN_LAYERS)
	D                    = make([][][]float64, 1+NUM_HIDDEN_LAYERS)
	X                    = []string{
		"http://puu.sh/pPVbl/da020edbb7.jpg",
		"http://puu.sh/pQ0pc/eb70ad235b.jpg",
		"http://puu.sh/pQ7r4/e652a1b1e7.png",
		"http://puu.sh/pQ7jw/3a97015b82.png",
		"http://puu.sh/pQ7AG/51317b5b1d.jpg",
		"http://puu.sh/pQ7UG/20a18b5b0b.jpg",
	}
	x = [][]float64{}
	y = [][]float64{
		S,
		A,
		A,
		B,
		B,
		B,
	}
)

func classify(url string) string {

	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)

	rand.Seed(time.Now().UTC().UnixNano())
	x = [][]float64{}
	var img image.Image
	for _, v := range X {
		img = getImageFromURL(v)
		z := 0
		x_new := make([]float64, 1200)
		for i := 0; i < 20; i++ {
			for j := 0; j < 20; j++ {
				r, g, b, _ := img.At(i, j).RGBA()
				x_new[z] = float64(r) / 655356.0
				z++
				x_new[z] = float64(g) / 655356.0
				z++
				x_new[z] = float64(b) / 655356.0
				z++
			}
		}
		x = append(x, x_new)
	}

	img2 := getImageFromURL(url)
	//Now let's set up the neural network
	//We have 1200 nodes in the input layer: 20x20x3
	//that's three nodes per pixel, holding r, g, b, values
	//now we need a matrix for weights
	//in general, we have 2 + NUM_HIDDEN_LAYERS layers
	//and we need weights going from a layer i to i + 1
	//i = {1, 2, ..., 1 + NUM_HIDDEN_LAYERS}
	//so in total we have a (1 + NUM_HIDDEN_LAYERS) sized array of matrices
	//this is a HIDDEN_LAYER_SIZE x HIDDEN_LAYER_SIZE matrix between two hidden,
	//use INPUT_LAYER_SIZE and OUTPUT_LAYER_SIZE accordingly
	for i := 0; i < NUM_HIDDEN_LAYERS; i++ {
		a[i] = make([]float64, HIDDEN_LAYER_SIZE)
	}
	w[0] = make([][]float64, INPUT_LAYER_SIZE)
	new_w[0] = make([][]float64, INPUT_LAYER_SIZE)
	for i := 0; i < INPUT_LAYER_SIZE; i++ {
		w[0][i] = make([]float64, HIDDEN_LAYER_SIZE)
		new_w[0][i] = make([]float64, HIDDEN_LAYER_SIZE)
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			w[0][i][j] = rand.Float64() / 100000.0
		}
	}
	for i := 1; i < NUM_HIDDEN_LAYERS; i++ {
		w[i] = make([][]float64, HIDDEN_LAYER_SIZE)
		new_w[i] = make([][]float64, HIDDEN_LAYER_SIZE)
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			w[i][j] = make([]float64, HIDDEN_LAYER_SIZE)
			new_w[i][j] = make([]float64, HIDDEN_LAYER_SIZE)
			for k := 0; k < HIDDEN_LAYER_SIZE; k++ {
				w[i][j][k] = rand.Float64() / 100000.0
			}
		}
	}
	w[NUM_HIDDEN_LAYERS] = make([][]float64, HIDDEN_LAYER_SIZE)
	new_w[NUM_HIDDEN_LAYERS] = make([][]float64, HIDDEN_LAYER_SIZE)
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		w[NUM_HIDDEN_LAYERS][i] = make([]float64, OUTPUT_LAYER_SIZE)
		new_w[NUM_HIDDEN_LAYERS][i] = make([]float64, OUTPUT_LAYER_SIZE)
		for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
			w[NUM_HIDDEN_LAYERS][i][j] = rand.Float64() / 100000.0
		}
	}

	delta[1+NUM_HIDDEN_LAYERS] = make([]float64, OUTPUT_LAYER_SIZE)
	for i := 1; i < 1+NUM_HIDDEN_LAYERS; i++ {
		delta[i] = make([]float64, HIDDEN_LAYER_SIZE)
	}

	//Delta, D should be same size as w
	for l := 0; l < len(w); l++ {
		Delta[l] = make([][]float64, len(w[l]))
		D[l] = make([][]float64, len(w[l]))
		for i := 0; i < len(w[l]); i++ {
			Delta[l][i] = make([]float64, len(w[l][i]))
			D[l][i] = make([]float64, len(w[l][i]))
		}
	}
	last_cost := 1000000.0
	for qq := 0; qq > -1; qq++ {
		fmt.Printf("Iteration: %v; cost: %v\n", qq, last_cost)
		if last_cost-J(x, y) <= 0.001 {
			break
		}
		last_cost = J(x, y)
		backprop()
		for i := 0; i < INPUT_LAYER_SIZE; i++ {
			for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
				new_w[0][i][j] = w[0][i][j] - LEARNING_RATE*D[0][i][j]
			}
		}
		for l := 1; l < NUM_HIDDEN_LAYERS; l++ {
			for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
				for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
					new_w[l][i][j] = w[l][i][j] - LEARNING_RATE*D[l][i][j]
				}
			}
		}
		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
				new_w[NUM_HIDDEN_LAYERS][i][j] = w[NUM_HIDDEN_LAYERS][i][j] - LEARNING_RATE*D[NUM_HIDDEN_LAYERS][i][j]
			}
		}
		for i := 0; i < INPUT_LAYER_SIZE; i++ {
			for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
				w[0][i][j] = new_w[0][i][j]
			}
		}
		for l := 1; l < NUM_HIDDEN_LAYERS; l++ {
			for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
				for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
					w[l][i][j] = new_w[l][i][j]
				}
			}
		}
		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
				w[NUM_HIDDEN_LAYERS][i][j] = new_w[NUM_HIDDEN_LAYERS][i][j]
			}
		}
	}
	fmt.Printf("%v", url)

	z := 0
	x_new := make([]float64, 1200)
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			r, g, b, _ := img2.At(i, j).RGBA()
			x_new[z] = float64(r)
			z++
			x_new[z] = float64(g)
			z++
			x_new[z] = float64(b)
			z++
		}
	}
	op := output(x_new)
	tiers := []string{"S tier", "A tier", "shit tier"}
	mv := -1.0
	mi := 0
	for i := 0; i < len(op); i++ {
		fmt.Printf("%v\n", op[i])
		if op[i] > mv {
			mv = op[i]
			mi = i
		}
	}
	return fmt.Sprintf("Your waifu is %v\n", tiers[mi])
}

func backprop() {
	for i := 0; i < len(x); i++ {
		op := output(x[i])
		for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
			delta[1+NUM_HIDDEN_LAYERS][j] = op[j] - y[i][j]
		}
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			dot := 0.0
			for k := 0; k < OUTPUT_LAYER_SIZE; k++ {
				dot += w[NUM_HIDDEN_LAYERS][j][k] * delta[1+NUM_HIDDEN_LAYERS][k]
			}
			delta[NUM_HIDDEN_LAYERS][j] = dot * a[NUM_HIDDEN_LAYERS-1][j] * (1 - a[NUM_HIDDEN_LAYERS-1][j])
		}
		for l := NUM_HIDDEN_LAYERS - 1; l >= 1; l-- {
			for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
				dot := 0.0
				for k := 0; k < HIDDEN_LAYER_SIZE; k++ {
					dot += w[l][j][k] * delta[l+1][k]
				}
				delta[l][j] = dot * a[l-1][j] * (1 - a[l-1][j])
			}
		}
		for I := 0; I < INPUT_LAYER_SIZE; I++ {
			for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
				Delta[0][I][j] += x[i][I] * delta[1][j]
			}
		}
		for l := 1; l < NUM_HIDDEN_LAYERS; l++ {
			for I := 0; I < HIDDEN_LAYER_SIZE; I++ {
				for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
					Delta[l][I][j] += a[l-1][I] * delta[l+1][j]
				}
			}
		}
		for I := 0; I < HIDDEN_LAYER_SIZE; I++ {
			for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
				Delta[NUM_HIDDEN_LAYERS][I][j] += a[NUM_HIDDEN_LAYERS-1][I] * delta[1+NUM_HIDDEN_LAYERS][j]
			}
		}
	}
	for i := 0; i < INPUT_LAYER_SIZE; i++ {
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			D[0][i][j] = 1.0/float64(len(x))*Delta[0][i][j] + LAMBDA*w[0][i][j]
		}
	}

	for l := 1; l < NUM_HIDDEN_LAYERS; l++ {
		for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
			for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
				D[l][i][j] = 1.0/float64(len(x))*Delta[l][i][j] + LAMBDA*w[l][i][j]
			}
		}
	}

	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		for j := 0; j < OUTPUT_LAYER_SIZE; j++ {
			D[NUM_HIDDEN_LAYERS][i][j] = 1.0/float64(len(x))*Delta[NUM_HIDDEN_LAYERS][i][j] + LAMBDA*w[NUM_HIDDEN_LAYERS][i][j]
		}
	}
}

func J(x [][]float64, y [][]float64) float64 {
	m := len(y)
	cost := 0.0
	for i := 0; i < m; i++ {
		op := output(x[i])
		for k := 0; k < OUTPUT_LAYER_SIZE; k++ {
			cost += y[i][k]*math.Log(op[k]) + (1-y[i][k])*math.Log(1-op[k])
		}
	}
	r := 0.0
	for i := 0; i < len(w); i++ {
		for j := 0; j < len(w[i]); j++ {
			for k := 0; k < len(w[i][j]); k++ {
				r += w[i][j][k] * w[i][j][k]
			}
		}
	}
	r *= LAMBDA / (2.0 * float64(m))
	return -1.0/float64(m)*cost + r
}

func output(input []float64) []float64 {
	for i := 0; i < HIDDEN_LAYER_SIZE; i++ {
		for j := 0; j < INPUT_LAYER_SIZE; j++ {
			a[0][i] += w[0][j][i] * input[j]
		}
		a[0][i] = g(a[0][i])
	}

	for i := 1; i < NUM_HIDDEN_LAYERS; i++ {
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			for k := 0; k < HIDDEN_LAYER_SIZE; k++ {
				a[i][j] += w[i][k][j] * a[i-1][k]
			}
			a[i][j] = g(a[i][j])
		}
	}

	op := make([]float64, OUTPUT_LAYER_SIZE)

	for i := 0; i < OUTPUT_LAYER_SIZE; i++ {
		for j := 0; j < HIDDEN_LAYER_SIZE; j++ {
			op[i] += w[NUM_HIDDEN_LAYERS][j][i] * a[NUM_HIDDEN_LAYERS-1][j]
		}
		op[i] = g(op[i])
	}
	return op
}

func g(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(x))
}

func getImageFromURL(url string) image.Image {
	fmt.Printf("%v\n", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()

	m, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	bounds := m.Bounds()

	X1 := bounds.Min.X
	Y1 := bounds.Min.Y
	X2 := bounds.Max.X
	Y2 := bounds.Max.Y

	w := X2 - X1
	h := Y2 - Y1

	vCompressionRatio := h / 20
	hCompressionRatio := w / 20
	vOverflow := h % 20
	hOverflow := w % 20

	var r image.Rectangle
	var z image.Point
	z.X = 0
	z.Y = 0
	r.Min = z
	z.X = 20
	z.Y = 20
	r.Max = z
	img := image.NewNRGBA(r)
	for x0 := 0; x0 < 20; x0++ {
		for y0 := 0; y0 < 20; y0++ {
			x1 := x0*hCompressionRatio + min(hOverflow, x0)
			y1 := y0*vCompressionRatio + min(vOverflow, y0)
			x2 := x1 + hCompressionRatio - 1
			y2 := y1 + vCompressionRatio - 1
			if x0 < hOverflow {
				x2++
			}
			if y0 < vOverflow {
				y2++
			}
			R := 0
			G := 0
			B := 0
			A := 0
			for x := x1; x <= x2; x++ {
				for y := y1; y <= y2; y++ {
					r, g, b, a := m.At(x+x0, y+y0).RGBA()
					R += int(r / 256)
					G += int(g / 256)
					B += int(b / 256)
					A += int(a / 256)
				}
			}
			R /= ((x2 - x1 + 1) * (y2 - y1 + 1))
			G /= ((x2 - x1 + 1) * (y2 - y1 + 1))
			B /= ((x2 - x1 + 1) * (y2 - y1 + 1))
			A /= ((x2 - x1 + 1) * (y2 - y1 + 1))
			var k color.NRGBA
			k.R = uint8(R)
			k.G = uint8(G)
			k.B = uint8(B)
			k.A = uint8(A)
			img.SetNRGBA(x0, y0, color.NRGBA(k))
		}
	}
	//writer, _ := os.Create("compressed.png")
	//png.Encode(writer, img)
	fmt.Printf("Read successfully\n")
	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
