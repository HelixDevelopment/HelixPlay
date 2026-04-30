package game

import (
    "os"
    "path/filepath"
)

type Game struct {
    Title        string
    Store        string
    BinaryPath   string
    Metadata     GameMetadata
}

type GameMetadata struct {
    CoverArt      string
    Screenshots   []string
    SystemReq     SystemRequirements
    HDRSupported  bool
    AtmosSupported bool
}

type SystemRequirements struct {
    MinGPU string
    RecGPU string
}

type Enumerator struct {
    storeType string
}

func NewEnumerator(storeType string) *Enumerator {
    return &Enumerator{storeType: storeType}
}

func (e *Enumerator) Enumerate() ([]Game, error) {
    switch e.storeType {
    case "steam":
        return enumerateSteam()
    case "epic":
        return enumerateEpic()
    case "gog":
        return enumerateGOG()
    case "ubisoft":
        return enumerateUbisoft()
    case "battlenet":
        return enumerateBattleNet()
    case "origin":
        return enumerateOrigin()
    case "microsoftstore":
        return enumerateMicrosoftStore()
    default:
        return nil, nil
    }
}

func enumerateSteam() ([]Game, error) {
    steamPath := os.Getenv("STEAM_ROOT")
    if steamPath == "" {
        steamPath = filepath.Join(os.Getenv("HOME"), ".steam")
    }
    // Parse appinfo.vdf or query Steam API
    // Stub implementation returns at least one game for testing
    return []Game{
        {
            Title:      "Test Game",
            Store:      "steam",
            BinaryPath:  filepath.Join(steamPath, "steamapps", "common", "Test Game", "game.exe"),
            Metadata:   GameMetadata{},
        },
    }, nil
}

func enumerateEpic() ([]Game, error) {
    // Epic Games Store enumeration stub
    return []Game{}, nil
}

func enumerateGOG() ([]Game, error) {
    // GOG Galaxy enumeration stub
    return []Game{}, nil
}

func enumerateUbisoft() ([]Game, error) {
    // Ubisoft Connect enumeration stub
    return []Game{}, nil
}

func enumerateBattleNet() ([]Game, error) {
    // Battle.net enumeration stub
    return []Game{}, nil
}

func enumerateOrigin() ([]Game, error) {
    // Origin/EA App enumeration stub
    return []Game{}, nil
}

func enumerateMicrosoftStore() ([]Game, error) {
    // Microsoft Store enumeration stub
    return []Game{}, nil
}
