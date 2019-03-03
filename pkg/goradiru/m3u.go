package goradiru

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"

	"github.com/grafov/m3u8"
)

func downloadEpisode(episode *Episode) (err error) {
	config := GetConfig()

	m3u8Url := getM3u8MasterPlaylist(episode.Url)

	progDir := config.ProgDir
	fileType := config.FileType
	filename := fmtFileName(episode, progDir, fileType)

	if fileType == "mp4" {
		err = convertM3u8ToMP4(m3u8Url, filename)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Not implemented such file type:" + fileType)
	}
	return nil
}

func convertM3u8ToMP4(masterM3u8Path string, filename string) error {
	f, err := newFFMPEG(masterM3u8Path)
	if err != nil {
		return err
	}

	f.setArgs(
		"-protocol_whitelist", "file,crypto,http,https,tcp,tls",
		"-movflags", "faststart",
		"-c", "copy",
		"-y",
		"-bsf:a", "aac_adtstoasc",
	)

	_, err = f.execute(filename + ".mp4")
	if err != nil {
		return err
	}
	return nil
}

//func convertMP4ToMP3(mp4path string, title string) error {
//	f, err := newFFMPEG(mp4path)
//	if err != nil {
//		return err
//	}
//
//	f.setArgs(
//		"-y",
//		"-acodec", "libmp3lame",
//		"-ab", "256k",
//	)
//
//	var name = title + ".mp3"
//	fmt.Println(name)
//
//	_, err = f.execute(name)
//	return err
//}

func getM3u8MasterPlaylist(m3u8FilePath string) string {
	resp, err := http.Get(m3u8FilePath) // nolint: gosec
	if err != nil {
		log.Fatal(err)
	}
	f := resp.Body

	p, t, err := m3u8.DecodeFrom(f, true)
	if err != nil {
		log.Fatal(err)
	}

	if t != m3u8.MASTER {
		log.Fatalf("not support file type [%v]", t)
	}

	return p.(*m3u8.MasterPlaylist).Variants[0].URI
}

//　出力ファイル名のフルパスを返す
func fmtFileName(episode *Episode, baseDir string, fileType string) string {
	dirname := filepath.Join(".", baseDir)
	filename := episode.Program.Title + "_" + episode.Series.Title + "_" + episode.Title + "." + fileType
	return filepath.Join(dirname, filename)
}
