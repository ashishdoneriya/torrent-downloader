package main

import (
	"bufio"
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
	"github.com/jlaffaye/ftp"
)

var args struct {
	torrentsDir string
	ftpHost     string
	ftpPort     string
	ftpUsername string
	ftpPassword string
}

func parseArguments() {
	args.torrentsDir = os.Args[1]
	args.ftpHost = os.Args[2]
	args.ftpPort = os.Args[3]
	args.ftpUsername = os.Args[4]
	args.ftpPassword = os.Args[5]
}

func createFtpClient() *ftp.ServerConn {
	client, err := ftp.Dial(args.ftpHost + ":" + args.ftpPort)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if err := client.Login(args.ftpUsername, args.ftpPassword); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return client
}

func main() {
	parseArguments()
	torrentFilesArr := getFiles(args.torrentsDir)
	totalTorrentsFiles := len(torrentFilesArr)
	for i := 0; i < totalTorrentsFiles; i += 5 {
		startIndex := i
		endIndex := i + 5
		if endIndex > totalTorrentsFiles {
			endIndex = totalTorrentsFiles
		}
		downloadTorrents(torrentFilesArr[startIndex:endIndex])
	}
}

func downloadTorrents(torrentFilesArr []os.FileInfo) {
	torrentClient, error := torrent.NewClient(nil)
	defer torrentClient.Close()
	if error != nil {
		fmt.Println("Unable to create torrent client. " + error.Error())
		return
	}

	downloadedFiles := list.New()
	for _, torrent := range torrentFilesArr {
		torrentObj, error := torrentClient.AddTorrentFromFile(args.torrentsDir + torrent.Name())
		if error != nil {
			fmt.Println("Unable to create torrent object. " + error.Error())
			return
		}
		<-torrentObj.GotInfo()
		torrentObj.DownloadAll()
		downloadedFilesTemp := torrentObj.Files()
		for _, file := range downloadedFilesTemp {
			downloadedFiles.PushBack(file.Path())
		}
	}
	torrentClient.WaitAll()
	transferDownloadedFiles(downloadedFiles)
}

func transferDownloadedFiles(downloadedFiles *list.List) {
	ftpClient := createFtpClient()
	if ftpClient != nil {
		return
	}

}

func getFiles(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func transfer(downloadedFiles []*torrent.File) {
	client, err := ftp.Dial("ftphost:port")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := client.Login("username", "password"); err != nil {
		fmt.Println(err.Error())
		return
	}

	var todel string
	for _, downloadedFile := range downloadedFiles {
		todel = filepath.Dir(downloadedFile.Path())
		client.MakeDir(filepath.Dir(downloadedFile.Path()))
		absDownloadedFilePath := downloadedFile.Path()
		file, err := os.Open(absDownloadedFilePath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		client.Stor(absDownloadedFilePath, bufio.NewReader(file))
		os.Remove(absDownloadedFilePath)
	}
	os.Remove(todel)
}
