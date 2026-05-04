package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	urlPostal = "https://github.com/open-mexico/sepomex-db-generator/releases/download/v1.1.1/db_postal.sqlite.zip"
	urlGeo    = "https://github.com/open-mexico/sepomex-db-generator/releases/download/v1.1.1/db_geo.sqlite.zip"
)

func main() {
	usarGeo := flag.Bool("geo", true, "Descargar BD con polígonos GeoJSON")
	flag.Parse()

	urlDescarga := urlPostal
	if *usarGeo {
		urlDescarga = urlGeo
	}

	fmt.Println("⏳ Descargando base de datos...")
	if err := descargarArchivo(urlDescarga, "temp.zip"); err != nil {
		fmt.Printf("❌ Error al descargar: %v\n", err)
		return
	}

	fmt.Println("🗜️ Descomprimiendo...")
	if err := extraerYRenombrar("temp.zip", "mapa.db"); err != nil {
		fmt.Printf("❌ Error al extraer: %v\n", err)
		return
	}
	_ = os.Remove("temp.zip")
	fmt.Println("✅ Base de datos lista en 'mapa.db'")
}

func descargarArchivo(url, destino string) (err error) {
	out, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extraerYRenombrar(rutaZip, destino string) (err error) {
	r, err := zip.OpenReader(rutaZip)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	if len(r.File) == 0 {
		return fmt.Errorf("zip vacío")
	}

	f, err := r.File[0].Open()
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	out, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, r.File[0].Mode())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, f)
	return err
}
