package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

// metaregBenchCmd represents the metaregBench command
var metaregBenchCmd = &cobra.Command{
	Use: "metaregBench",
	Run: func(cmd *cobra.Command, args []string) {
		// setupGui()
	},
}

type tt struct {
	b int64
	t int64
}

func (t tt) GetBlockHash() []byte {
	return []byte(strconv.FormatInt(t.b, 10))
}

func (t tt) GetTransactionHash() []byte {
	return []byte(strconv.FormatInt(t.t, 10))
}

/*func setupGui() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	g3 := ui.NewPar("")
	//g3.Width = 100
	g3.Height = 3
	g3.TextFgColor = ui.ColorGreen

	spl := ui.NewSparkline()
	spl.LineColor = ui.ColorRed
	spl.Height = 8

	splM := ui.NewSparklines(spl)
	splM.Width = 20
	splM.Height = 11

	hashrate := ui.NewPar("")
	hashrate.TextFgColor = ui.ColorRed
	hashrate.Height = 3
	hashrate.BorderLabel = "rate"

	since := ui.NewPar("")
	since.TextFgColor = ui.ColorBlue
	since.Height = 3
	since.BorderLabel = "time"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, splM),
		),
		ui.NewRow(
			ui.NewCol(2, 0, hashrate),
			ui.NewCol(8, 0, g3),
			ui.NewCol(2, 0, since),
		),
		ui.NewRow(),
	)
	ui.Body.Align()

	var last uint64

	storage, ids := setupDB()

	var rec1, rec2 model.KeySchema
	registry := &types.Registry{
		Name:      model.KeySchema{}.ModelName(),
		Ecosystem: &types.Ecosystem{Name: "aaa"},
	}

	t := tt{b: 2}
	go func() {
		for {
			tx := storage.Begin()
			lid := len(ids)
			for pos, id := range ids {
				t.t = int64(atomic.AddUint64(&last, 1))

				if pos == lid-1 {
					break
				}

				checkErr(tx.Get(registry, strconv.FormatInt(id, 10), &rec1))
				checkErr(tx.Get(registry, strconv.FormatInt(ids[pos+1], 10), &rec2))

				rec2.Amount = rec1.Amount
				rec1.Amount = "0"

				checkErr(tx.Update(
					t,
					registry,
					strconv.FormatInt(id, 10),
					rec1,
				))
				checkErr(tx.Update(
					t,
					registry,
					strconv.FormatInt(id, 10),
					rec2,
				))
			}

			t.b++
			if t.b > 5 {
				checkErr(tx.CleanBlockState([]byte(strconv.FormatInt(t.b-5, 10))))
			}
			checkErr(tx.Commit())
		}
	}()

	go func() {
		var buf uint64
		retro := make([]int, 1)
		start := time.Now()
		for range time.NewTicker(time.Second).C {
			l := atomic.LoadUint64(&last)
			if l > 0 {
				cur := l - buf
				buf = l
				if len(retro) < 30 {
					retro = append(retro, int(cur))
				} else {
					retro = append(retro[1:], int(cur))
				}

				splM.Lines[0].Data = retro
				hashrate.Text = fmt.Sprintf("%d tx/s", cur)
				since.Text = fmt.Sprintf("%ss", strconv.FormatFloat(time.Since(start).Seconds(), 'f', 0, 64))
				g3.Text = fmt.Sprintf("%d money transactions done", l)

				ui.Render(g3, hashrate, since, splM)
			}
		}
	}()

	ui.Handle("q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Loop()
}

func checkErr(err error) {
	if err != nil {
		ui.StopLoop()
		ui.Close()
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupDB() (types.MetadataRegistryStorage, []int64) {
	rollbacks := true
	persist := true
	pricing := true

	checkErr(os.RemoveAll("metabench.db"))
	db, err := memdb.OpenDB("metabench.db", persist)
	checkErr(err)

	storage, err := metadata.NewStorage(&kv.DatabaseAdapter{Database: *db}, []types.Index{
		{
			Registry: &types.Registry{Name: model.KeySchema{}.ModelName()},
			Name:     "amount",
			SortFn: func(a, b string) bool {
				return gjson.Get(b, "amount").Less(gjson.Get(a, "amount"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: model.Ecosystem{}.ModelName(), Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}, rollbacks, pricing)
	checkErr(err)

	t := tt{b: 0}
	metadataTx := storage.Begin()

	t.t++
	checkErr(metadataTx.Insert(
		t,
		&types.Registry{Name: model.Ecosystem{}.ModelName()},
		"aaa",
		model.Ecosystem{Name: "aaa"},
	))

	ids := make([]int64, 0)
	for i := 0; i < 10000; i++ {
		reg := types.Registry{
			Name:      model.KeySchema{}.ModelName(),
			Ecosystem: &types.Ecosystem{Name: "aaa"},
		}

		t.t = int64(i)
		id := rand.Int63()
		strId := strconv.FormatInt(id, 10)
		err := metadataTx.Insert(
			t,
			&reg,
			strId,
			model.KeySchema{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    "0",
			},
		)

		checkErr(err)
		ids = append(ids, id)
	}
	checkErr(metadataTx.Commit())
	return storage, ids
}
*/
