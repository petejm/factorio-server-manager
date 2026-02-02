package factorio

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ModPortalStruct struct {
	DownloadsCount int    `json:"downloads_count"`
	Name           string `json:"name"`
	Owner          string `json:"owner"`
	Releases       []struct {
		DownloadURL string `json:"download_url"`
		FileName    string `json:"file_name"`
		InfoJSON    struct {
			FactorioVersion Version `json:"factorio_version"`
		} `json:"info_json"`
		ReleasedAt    time.Time `json:"released_at"`
		Sha1          string    `json:"sha1"`
		Version       Version   `json:"version"`
		Compatibility bool      `json:"compatibility"`
	} `json:"releases"`
	Summary string `json:"summary"`
	Title   string `json:"title"`
}

// get all mods uploaded to the factorio modPortal
func ModPortalList() (interface{}, error, int) {
	req, err := http.NewRequest(http.MethodGet, "https://mods.factorio.com/api/mods?page_size=max", nil)
	if err != nil {
		return "error", err, http.StatusInternalServerError
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "error", err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	text, err := io.ReadAll(resp.Body)
	if err != nil {
		return "error", err, http.StatusInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(text)), resp.StatusCode
	}

	var jsonVal interface{}
	err = json.Unmarshal(text, &jsonVal)
	if err != nil {
		return "error", err, http.StatusInternalServerError
	}

	return jsonVal, nil, resp.StatusCode
}

// get the details (mod-info, releases, etc.) from a specific mod from the modPortal
func ModPortalModDetails(modId string) (ModPortalStruct, error, int) {
	var mod ModPortalStruct

	req, err := http.NewRequest(http.MethodGet, "https://mods.factorio.com/api/mods/"+modId, nil)
	if err != nil {
		return mod, err, http.StatusInternalServerError
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mod, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	text, err := io.ReadAll(resp.Body)
	if err != nil {
		return mod, err, http.StatusInternalServerError
	}

	err = json.Unmarshal(text, &mod)
	if err != nil {
		return mod, err, http.StatusInternalServerError
	}

	if resp.StatusCode != http.StatusOK {
		return ModPortalStruct{}, errors.New(string(text)), resp.StatusCode
	}

	server := GetFactorioServer()

	installedBaseVersion := Version{}
	if err := installedBaseVersion.UnmarshalText([]byte(server.BaseModVersion)); err != nil {
		log.Printf("error parsing base mod version: %s", err)
	}
	requiredVersion := NilVersion

	for key, release := range mod.Releases {
		requiredVersion = release.InfoJSON.FactorioVersion
		release.Compatibility = installedBaseVersion.Compatible(requiredVersion, ">=")
		mod.Releases[key] = release
	}

	return mod, nil, resp.StatusCode
}

//Log the user into factorio, so mods can be downloaded
func FactorioLogin(username string, password string) (error, int) {
	var err error

	resp, err := http.PostForm("https://auth.factorio.com/api-login",
		url.Values{"require_game_ownership": {"true"}, "username": {username}, "password": {password}})

	if err != nil {
		return err, http.StatusInternalServerError
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	bodyString := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return errors.New(bodyString), resp.StatusCode
	}

	// Try parsing as array first (old API format)
	var arrayResponse []string
	if err = json.Unmarshal(bodyBytes, &arrayResponse); err == nil && len(arrayResponse) >= 2 {
		credentials := Credentials{
			Username: arrayResponse[0],
			Userkey:  arrayResponse[1],
		}
		if err = credentials.Save(); err != nil {
			return err, http.StatusInternalServerError
		}
		return nil, http.StatusOK
	}

	// Try parsing as object (new API format)
	var objectResponse struct {
		Token   string `json:"token"`
		Username string `json:"username"`
	}
	if err = json.Unmarshal(bodyBytes, &objectResponse); err == nil && objectResponse.Token != "" {
		credentials := Credentials{
			Username: objectResponse.Username,
			Userkey:  objectResponse.Token,
		}
		if err = credentials.Save(); err != nil {
			return err, http.StatusInternalServerError
		}
		return nil, http.StatusOK
	}

	log.Printf("Unexpected auth response format: %s", bodyString)
	return errors.New("unexpected response format from auth API"), http.StatusInternalServerError
}
