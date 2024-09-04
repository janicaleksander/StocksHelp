package charts

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/google/uuid"
	"io"
	"os"
)

type KlineData struct {
	Date string
	Data [4]float64
}

func klineBase(name string, kd []KlineData) *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	for i := 0; i < len(kd); i++ {
		x = append(x, kd[i].Date)
		y = append(y, opts.KlineData{Value: kd[i].Data})
	}

	kline.SetGlobalOptions(

		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(true),
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	kline.SetXAxis(x).AddSeries(name, y)
	return kline
}

type KlineExamples struct{}

func (KlineExamples) Examples(name string, data []KlineData, id uuid.UUID) {
	page := components.NewPage()
	page.AddCharts(
		klineBase(name, data),
	)
	url := fmt.Sprintf("./static/chart/marketChart%v.html", id.String())
	f, err := os.Create(url)
	if err != nil {
		panic(err)

	}
	err = page.Render(io.MultiWriter(f))
	fmt.Println(err)
}
