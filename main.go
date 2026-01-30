package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
    "strings"
     securejoin "github.com/cyphar/filepath-securejoin"

)

var (
	version = "0.0.0"
	commit  = ""
)

func ExtractTarGz(gzipStream io.Reader, destDir string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.FromSlash(header.Name)
		name = strings.TrimLeft(name, string(filepath.Separator))
		if name == "" || name == "." {
			continue
		}
		
		outPath, err := securejoin.SecureJoin(destAbs, name)
		mode := header.FileInfo().Mode()

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(outPath, mode.Perm()); err != nil {
				return err
			}

		case tar.TypeReg:
			f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tarReader); err != nil {
				f.Close()
				return err
			}
			f.Close()

		default:
			return fmt.Errorf("tar: unknown type %q in %q", header.Typeflag, header.Name)
		}
	}
	return nil
}

func main() {
	log.SetFlags(0)

	exe := path.Base(os.Args[0])
	if len(os.Args) != 2 {
		log.Fatalln("Usage:", exe, "<input_tgz_file>")
	}

	if os.Args[1] == "--version" {
		fmt.Println(exe, version, commit)
		os.Exit(0)
	}

	r, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Close()

	if err := ExtractTarGz(r); err != nil {
		log.Fatalln(err)
	}
}
