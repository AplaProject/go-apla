package spawn

import (
	"errors"
	"fmt"
	"github.com/pb"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/go-thrust/lib/common"
)

func downloadFromUrl(url, filepath, version string) (fileName string, err error) {
	url = strings.Replace(url, "$V", version, 2)
	fileName = strings.Replace(filepath, "$V", version, 1)
	fmt.Println("Downloading 0", url, "to", fileName)

	output, err := os.Create(fileName)
	if err != nil {
		Log.Print("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Server return non-200 status: %v", response.Status))
		fmt.Println(err)
		return
	}

	// create bar
	bar := pb.New(int(response.ContentLength)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.Start()
	defer bar.Finish()

	// create multi writer
	writer := io.MultiWriter(output, bar)

	_, err = io.Copy(writer, response.Body)
	if err != nil {
		Log.Print("Error while downloading", url, "-", err)
		return
	}

	return
}
