package spawn

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cheggaaa/pb"

	. "github.com/miketheprogrammer/go-thrust/lib/common"
)

func downloadFromUrl(url, filepath, version string) (fileName string, err error) {
	url = strings.Replace(url, "$V", version, 2)
	fileName = strings.Replace(filepath, "$V", version, 1)
	if PathNotExist(fileName) == false {
		fmt.Println("Thrust already exists on filesystem .... skipping")
		return
	}
	fmt.Println("Extract directory was", GetDownloadPath())
	fmt.Println("Downloading", url, "to", fileName)

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
