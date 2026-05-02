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
	urlPostal = "https://github.com/open-mexico/sepomex-db-generator/releases/download/v1.0.0/db_postal.sqlite.zip"
	urlGeo    = "https://github.com/open-mexico/sepomex-db-generator/releases/download/v1.0.0/db_geo.sqlite.zip"
)

func main() {
	usarGeo := flag.Bool("geo", false, "Descargar BD con polígonos GeoJSON")
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
	os.Remove("temp.zip")
	fmt.Println("✅ Base de datos lista en 'mapa.db'")
}

func descargarArchivo(url, destino string) error {
	out, err := os.Create(destino)
	if err != nil { return err }
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil { return err }
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extraerYRenombrar(rutaZip, destino string) error {
	r, err := zip.OpenReader(rutaZip)
	if err != nil { return err }
	defer r.Close()

	if len(r.File) == 0 { return fmt.Errorf("zip vacío") }
	
	f, err := r.File[0].Open()
	if err != nil { return err }
	defer f.Close()

	out, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, r.File[0].Mode())
	if err != nil { return err }
	defer out.Close()

	_, err = io.Copy(out, f)
	return err
}