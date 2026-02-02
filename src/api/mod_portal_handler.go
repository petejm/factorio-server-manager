package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/OpenFactorioServerManager/factorio-server-manager/factorio"
	"github.com/gorilla/mux"
)

func ModPortalListModsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var statusCode int
	resp, err, statusCode = factorio.ModPortalList()
	w.WriteHeader(statusCode)
	if err != nil {
		resp = fmt.Sprintf("Error in listing mods from mod portal: %s\nresponse: %+v", err, resp)
		log.Println(resp)
		return
	}
}

// ModPortalModInfoHandler returns JSON response with the mod details
func ModPortalModInfoHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	vars := mux.Vars(r)
	modId := vars["mod"]

	var statusCode int
	resp, err, statusCode = factorio.ModPortalModDetails(modId)

	if err != nil {
		resp = fmt.Sprintf("Error in getting mod details from mod portal: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
}

func ModPortalInstallHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	// Get Data out of the request
	var data struct {
		DownloadURL string `json:"downloadUrl"`
		Filename    string `json:"fileName"`
		ModName     string `json:"modName"`
	}
	resp, err = ReadFromRequestBody(w, r, &data)
	if err != nil {
		return
	}

	mods, resp, err := CreateNewMods(w)
	if err != nil {
		return
	}

	err = mods.DownloadMod(data.DownloadURL, data.Filename, data.ModName)
	if err != nil {
		resp = fmt.Sprintf("Error downloading a mod: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp = mods.ListInstalledMods()
}

func ModPortalLoginHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	resp, err = ReadFromRequestBody(w, r, &data)
	if err != nil {
		return
	}

	err, statusCode := factorio.FactorioLogin(data.Username, data.Password)
	w.WriteHeader(statusCode)
	if err != nil {
		resp = fmt.Sprintf("Error trying to login into Factorio: %s", err)
		log.Println(resp)
		return
	}
}

func ModPortalLoginStatusHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	var credentials factorio.Credentials
	resp, err = credentials.Load()

	if err != nil {
		resp = fmt.Sprintf("Error getting the factorio credentials: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func ModPortalLogoutHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	var credentials factorio.Credentials
	err = credentials.Del()

	if err != nil {
		resp = fmt.Sprintf("Error on logging out of factorio: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp = false
}

func ModPortalInstallMultipleHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var data []struct {
		Name    string           `json:"name"`
		Version factorio.Version `json:"version"`
	}
	resp, err = ReadFromRequestBody(w, r, &data)
	if err != nil {
		return
	}

	log.Printf("InstallMultiple: received %d mods to install", len(data))
	for i, mod := range data {
		log.Printf("  [%d] %s @ %s", i, mod.Name, mod.Version)
	}

	if len(data) == 0 {
		log.Println("InstallMultiple: no mods to install")
		resp = []interface{}{}
		return
	}

	modList, resp, err := CreateNewMods(w)
	if err != nil {
		return
	}

	for _, datum := range data {
		// skip base mod because it is already included in factorio
		if datum.Name == "base" {
			continue
		}
		details, err, statusCode := factorio.ModPortalModDetails(datum.Name)
		if err != nil || statusCode != http.StatusOK {
			resp = fmt.Sprintf("Error in getting mod details from mod portal: %s", err)
			log.Println(resp)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//find correct mod-version
		var found = false
		for _, release := range details.Releases {
			if release.Version.Equals(datum.Version) {
				found = true

				err := modList.DownloadMod(release.DownloadURL, release.FileName, details.Name)
				if err != nil {
					resp = fmt.Sprintf("Error downloading mod {%s}, error: %s", details.Name, err)
					log.Println(resp)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				break
			}
		}
		if !found {
			resp = fmt.Sprintf("Error downloading mod {%s}: version %s not found on mod portal", details.Name, datum.Version)
			log.Println(resp)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Successfully installed %d mods", len(data))
	resp = modList.ListInstalledMods()
}

// ModPortalInstallByNameHandler installs mods by name, using the latest compatible version
func ModPortalInstallByNameHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp interface{}

	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var data []struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	resp, err = ReadFromRequestBody(w, r, &data)
	if err != nil {
		return
	}

	// Filter to only enabled mods
	var enabledMods []string
	for _, mod := range data {
		if mod.Enabled && mod.Name != "base" && mod.Name != "elevated-rails" && mod.Name != "quality" && mod.Name != "space-age" {
			enabledMods = append(enabledMods, mod.Name)
		}
	}

	log.Printf("InstallByName: received %d mods (%d enabled, excluding base game mods)", len(data), len(enabledMods))

	if len(enabledMods) == 0 {
		log.Println("InstallByName: no mods to install")
		resp = []interface{}{}
		return
	}

	modList, resp, err := CreateNewMods(w)
	if err != nil {
		return
	}

	server := factorio.GetFactorioServer()
	installedBaseVersion := factorio.Version{}
	if err := installedBaseVersion.UnmarshalText([]byte(server.BaseModVersion)); err != nil {
		log.Printf("error parsing base mod version: %s", err)
	}

	installedCount := 0
	for _, modName := range enabledMods {
		details, err, statusCode := factorio.ModPortalModDetails(modName)
		if err != nil || statusCode != http.StatusOK {
			log.Printf("Warning: could not get details for mod %s: %v", modName, err)
			continue // Skip mods we can't find instead of failing entirely
		}

		// Find the latest compatible release
		var bestRelease *struct {
			DownloadURL string
			FileName    string
			Version     factorio.Version
		}
		for _, release := range details.Releases {
			if release.Compatibility {
				if bestRelease == nil || release.Version.Greater(bestRelease.Version) {
					bestRelease = &struct {
						DownloadURL string
						FileName    string
						Version     factorio.Version
					}{
						DownloadURL: release.DownloadURL,
						FileName:    release.FileName,
						Version:     release.Version,
					}
				}
			}
		}

		if bestRelease == nil {
			log.Printf("Warning: no compatible version found for mod %s", modName)
			continue
		}

		log.Printf("Installing %s @ %s", modName, bestRelease.Version)
		err = modList.DownloadMod(bestRelease.DownloadURL, bestRelease.FileName, modName)
		if err != nil {
			log.Printf("Warning: error downloading mod %s: %s", modName, err)
			continue
		}
		installedCount++
	}

	log.Printf("Successfully installed %d of %d mods", installedCount, len(enabledMods))
	resp = modList.ListInstalledMods()
}
