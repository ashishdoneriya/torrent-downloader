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
	torrentClient.Close()
}

func transferDownloadedFiles(downloadedFiles *list.List) {
	ftpClient := createFtpClient()
	if ftpClient != nil {
		return
	}
	for temp := downloadedFiles.Front(); temp != nil; temp = temp.Next() {
		localFilePath := temp.Value.(string)
		transferFileToFTP(ftpClient, localFilePath)
		// Removing original downloaded file
		os.Remove(localFilePath)
	}
	ftpClient.Logout()
}

func transferFileToFTP(ftpClient *ftp.ServerConn, localFilePath string) {
	createRemoteDirRecur(ftpClient, localFilePath)
	file, err := os.Open(localFilePath)
	if err != nil {
		fmt.Println("Unable to transfer file " + localFilePath + " to ftp " + err.Error())
	}
	err1 := ftpClient.Stor(localFilePath, bufio.NewReader(file))
	if err1 != nil {
		fmt.Println("Unable to transfer file " + localFilePath + " to ftp " + err.Error())
	}
}

func createRemoteDirRecur(ftpClient *ftp.ServerConn, localFilePath string) {
	if localFilePath == "." || localFilePath == ".." || localFilePath == "" || localFilePath == "/" || localFilePath == "./" {
		return
	}
	createRemoteDirRecur(ftpClient, filepath.Dir(localFilePath))
	ftpClient.MakeDir(localFilePath)
}

func getFiles(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return files
}
