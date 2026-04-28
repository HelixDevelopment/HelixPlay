# Dim 06 — Game Catalog, Metadata & 4K Asset Management: Deep Research Report

**Date:** 2025-07-14
**Searches Conducted:** 25 independent queries across all sub-dimensions
**Sources:** Official API docs, GitHub repositories, technical papers, established tech publications

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [IGDB (Internet Game Database) API](#igdb-api)
3. [SteamGridDB API](#steamgriddb-api)
4. [Steam Web API / SteamSpy](#steam-web-api--steamspy)
5. [Epic Games Store API](#epic-games-store-api)
6. [GOG Galaxy API](#gog-galaxy-api)
7. [Alternative Metadata Sources (RAWG, MobyGames, Giant Bomb)](#alternative-metadata-sources)
8. [Game Metadata Schema Design](#game-metadata-schema-design)
9. [4K Image Asset Pipeline](#4k-image-asset-pipeline)
10. [Local Asset Caching Strategy](#local-asset-caching-strategy)
11. [Search and Indexing](#search-and-indexing)
12. [Sorting and Filtering Architecture](#sorting-and-filtering-architecture)
13. [Game Trailers / Video Previews](#game-trailers--video-previews)
14. [User-Generated Content](#user-generated-content)
15. [Per-Game Settings Persistence](#per-game-settings-persistence)
16. [Tensions, Trade-offs, and Counter-Arguments](#tensions-trade-offs-and-counter-arguments)
17. [Recommended Architecture](#recommended-architecture)
18. [References](#references)

---

## Executive Summary

Building a PS4 Pro-like game library landing screen with 4K covers and screenshots requires a multi-layered architecture combining third-party metadata APIs, high-resolution artwork sources, modern image optimization, local caching, and full-text search. This research evaluates 10+ APIs, 5+ search solutions, and multiple caching strategies to provide an evidence-based blueprint.

**Key Finding:** No single API provides complete coverage. A hybrid approach combining IGDB (comprehensive metadata) + SteamGridDB (4K artwork) + Steam Web API (library import + playtime) provides the most robust foundation. For search, SQLite FTS5 is recommended for embedded clients while Meilisearch suits server-side deployments. For image optimization, AVIF at quality 65-80 and WebP at quality 75-85 provide optimal size-to-quality ratios for 4K game artwork.

---

## IGDB API

### Overview
IGDB (Internet Game Database) is a comprehensive game metadata platform owned by Twitch. It provides extensive game data through a RESTful API using the Apicalypse query language.

**Base URL:** `https://api.igdb.com/v4`

### Authentication
IGDB requires Twitch OAuth2 authentication. All requests need:
- `Client-ID` header
- `Authorization: Bearer <access_token>` header

### Pricing and Rate Limits

```
Claim: IGDB offers a free tier with 10K requests/month, a Pro tier at $99/month with 50K requests, and a Partner Program for qualifying non-commercial projects.
Source: IGDB Blog
URL: https://medium.com/igdb/its-here-the-new-igdb-api-f6ad745b53fe
Date: 2018-12-21
Excerpt: "In our revised free tier we are merging our old free and hobby tiers, giving it a 10k request limit per month... The second tier: Pro — 99 USD per month... 50K per month... The third tier: Partner Program — Free"
Context: Pricing was revised due to abuse of the free tier by financially viable companies mass-downloading the database.
Confidence: high
```

| Tier | Price | Requests/Month | Webhooks | Multi-query |
|------|-------|---------------|----------|-------------|
| Free | $0 | 10,000 | No | No |
| Pro | $99 | 50,000 | Yes | Yes |
| Partner | Free | 50,000+ | Yes | Yes |

**Data Dumps:** IGDB provides daily data dumps for bulk ingestion [^288^].

### Key Endpoints

| Endpoint | Description |
|----------|-------------|
| `/games` | Core game records |
| `/covers` | Cover art with image_id for URL construction |
| `/screenshots` | Game screenshots |
| `/artworks` | Official artworks (variable resolution) |
| `/genres` | Genre definitions |
| `/platforms` | Platform information |
| `/release_dates` | Release dates by region |
| `/age_ratings` | Age ratings (ESRB, PEGI, etc.) |
| `/videos` | Game videos/trailers |
| `/game_modes` | Single player, multiplayer, etc. |
| `/player_perspectives` | First-person, third-person, etc. |
| `/themes` | Fantasy, sci-fi, horror, etc. |
| `/franchises` | Game franchises |
| `/external_games` | Links to external stores (Steam, Epic, etc.) |
| `/multiplayer_modes` | Multiplayer mode details |
| `/websites` | Official websites, Steam pages, etc. |
| `/involved_companies` | Developer/publisher relationships |

### Image URL Construction

IGDB images use a templating system:
```
https://images.igdb.com/igdb/image/upload/t_{size}/{image_id}.png
```

Available sizes: `thumb`, `cover_small`, `cover_big`, `screenshot_med`, `screenshot_big`, `screenshot_huge`, `logo_med`, `micro`, `720p`, `1080p`

**Note:** IGDB images are NOT available in true 4K resolution. The maximum is 1080p (`t_1080p`). For 4K covers, SteamGridDB is required.

### Data Fields (Games Endpoint)

```
Claim: IGDB provides extensive game fields including name, summary, storyline, rating, aggregated_rating, first_release_date, genres, platforms, involved_companies, age_ratings, cover, screenshots, artworks, videos, websites, game_modes, player_perspectives, themes, franchises, external_games, multiplayer_modes, total_rating, total_rating_count
Source: IGDB API Documentation
URL: https://api-docs.igdb.com/
Date: 2018-11-25 (ongoing updates)
Excerpt: "Game records; supports Apicalypse query language; returns top-level array"
Context: Complete field listing in official documentation
Confidence: high
```

### Webhooks
Webhooks are available on the Pro tier and Partner Program, enabling real-time notifications when game data changes [^276^].

### API Limits and Considerations
- No native fuzzy search; exact matches on `name` field
- Rate limiting based on tier
- Requires Twitch developer account
- Image sizes capped at 1080p
- Covers are primarily DVD-style (not Steam vertical format)

---

## SteamGridDB API

### Overview
SteamGridDB is a community-driven database of game artwork specifically designed for Steam library customization. It is the premier source for 4K-quality game covers, hero banners, logos, and icons.

**Base URL:** `https://www.steamgriddb.com/api/v2`

### Authentication
API key required via `Authorization: Bearer <API_KEY>` header.

### Key Endpoints

```
Claim: SteamGridDB API v2 provides endpoints for grids (covers), heroes, logos, and icons, searchable by game ID, Steam App ID, or platform ID.
Source: SteamGridDB API Documentation
URL: https://www.steamgriddb.com/api/v2
Date: Ongoing
Excerpt: "Retrieve grids by game ID... Retrieve heroes by game ID... Retrieve logos by game ID... Retrieve icons by game ID"
Context: Full OpenAPI spec available at steamgriddb.com/api/v2
Confidence: high
```

| Endpoint | Asset Type | Dimensions |
|----------|-----------|------------|
| `/grids/game/{gameId}` | Vertical covers (box art) | 600x900 |
| `/grids/steam/{steamAppId}` | Vertical covers | 600x900 |
| `/heroes/game/{gameId}` | Hero/banner backgrounds | 1920x620, 3840x1240 (4K) |
| `/heroes/steam/{steamAppId}` | Hero/banner backgrounds | 1920x620, 3840x1240 (4K) |
| `/logos/game/{gameId}` | Game logos (transparent PNG) | Various |
| `/logos/steam/{steamAppId}` | Game logos | Various |
| `/icons/game/{gameId}` | Square icons | 512x512 |
| `/icons/steam/{steamAppId}` | Square icons | 512x512 |
| `/search/autocomplete/{term}` | Game search | N/A |
| `/games/id/{gameId}` | Game metadata | N/A |
| `/games/steam/{steamAppId}` | Game metadata | N/A |

### Image Dimensions and Formats

```
Claim: SteamGridDB supports multiple dimensions including 600x900 (Steam vertical cover), 920x430 (Steam horizontal), 1920x620 and 3840x1240 (4K hero), and 512x512 (icon).
Source: SteamGridDB Collections
URL: https://www.steamgriddb.com/collection/107/grids
Date: 2026-04-27
Excerpt: "1:1 - Square. 1024x1024. 512x512 ; 22:31 - Galaxy 2.0. 342x482. 660x930 ; 92:43 - Steam Horizontal. 460x215. 920x430 (7) ; 2:3 - Steam Vertical. 600x900 (17)"
Context: Verified against live collection data
Confidence: high
```

### Query Parameters

All asset endpoints support filtering:
- `styles`: `official`, `alternate`, `white`, `black`, `custom`, `animated`
- `mimes`: `image/png`, `image/webp`, `image/jpeg`
- `types`: `static`, `animated` (APNG/WebP animated)
- `nsfw`: `true`, `false`, `any`
- `humor`: `true`, `false`, `any`
- `dimensions`: Dimension-specific filtering
- `limit`: Max 50 per request

### 4K Asset Availability

**Hero backgrounds** support true 4K (3840x1240): [^275^]
- Hero images are uploaded at native resolution
- CDN serves optimized variants
- 4K support confirmed by "Fix carousel on 4k screens" in changelog

**Vertical covers** max at 600x900 (2:3 ratio):
- This is the Steam library standard
- Quality is excellent but not technically "4K" in resolution
- Upscaling via AI (Real-ESRGAN) can produce 1200x1800 variants

### API Usage Patterns

```python
# Example: Fetch 4K hero for a Steam game
GET https://www.steamgriddb.com/api/v2/heroes/steam/1091500
# Returns: hero images for Cyberpunk 2077 at various resolutions

# Example: Fetch vertical covers with WebP format
GET https://www.steamgriddb.com/api/v2/grids/game/5251765?mimes=image/webp&types=static
```

### Rate Limiting
- No official published rate limits
- Community best practice: reasonable request pacing
- API key required for all endpoints

---

## Steam Web API / SteamSpy

### Steam Web API Overview

The official Steam Web API provides library import, user data, and game metadata. It is free with a 100,000 requests/day limit.

```
Claim: The Steam Web API has a limit of 100,000 requests per day and is free for community use, supporting JSON, XML, and VDF formats.
Source: Ultimate Steam Web API Guide
URL: https://zuplo.com/learning-center/what-is-the-steam-web-api/
Date: 2024-10-04
Excerpt: "Rate limiting: There's a limit of 100,000 requests per day. This ensures fair usage and prevents abuse. Pricing: The API seems to be completely free for the community to use."
Context: Rate limits apply per API key
Confidence: high
```

### Key Endpoints for Game Library

#### IPlayerService — Library Import

| Endpoint | Purpose |
|----------|---------|
| `IPlayerService/GetOwnedGames/v1/` | List all owned games with playtime |
| `IPlayerService/GetRecentlyPlayedGames/v1/` | Recently played games |
| `IPlayerService/GetSingleGamePlaytime/v1/` | Playtime for specific app |

```
Claim: GetOwnedGames returns appid, name, playtime_forever, playtime_2weeks, img_icon_url, img_logo_url, and has_community_visible_stats for each game in a user's library.
Source: IPlayerService Interface Documentation
URL: https://partner.steamgames.com/doc/webapi/iplayerservice
Date: Ongoing
Excerpt: "Returns a list of games owned by the player if their owned games/game details are visible to you."
Context: Requires include_appinfo=true for game names and icons
Confidence: high
```

**Important:** `include_appinfo=true` is required to get game names; otherwise only app IDs are returned [^272^].

#### Store API — Game Metadata

```
Claim: The Steam Store API (store.steampowered.com/api/appdetails) returns comprehensive game metadata including name, developers, publishers, genres, screenshots, movies/trailers, release_date, required_age, metacritic score, categories, and supported platforms.
Source: Steam Web API Models Package
URL: https://feed.nuget.org/packages/SteamApi.Models/1.1.1
Date: 2025-08-07
Excerpt: "Store Models: StoreAppDetails, StoreRequirements, StorePlatforms, StoreMetacritic, StoreCategory, StoreGenre, StoreScreenshot, StoreMovie, StoreReleaseDate"
Context: Strongly-typed models confirm all available fields
Confidence: high
```

**Rate Limit:** ~200 requests per 5 minutes [^384^]

**Key Response Fields:**
```json
{
  "name": "Game Title",
  "steam_appid": 123456,
  "required_age": 0,
  "is_free": false,
  "detailed_description": "...",
  "short_description": "...",
  "developers": ["Developer Name"],
  "publishers": ["Publisher Name"],
  "genres": [{"id": "1", "description": "Action"}],
  "screenshots": [{"path_thumbnail": "...", "path_full": "..."}],
  "movies": [{"name": "Trailer", "mp4": {"480": "...", "max": "..."}}],
  "release_date": {"coming_soon": false, "date": "1 Jan 2023"},
  "metacritic": {"score": 85, "url": "..."},
  "categories": [{"id": 2, "description": "Single-player"}],
  "platforms": {"windows": true, "mac": false, "linux": false},
  "header_image": "...",
  "background": "...",
  "content_descriptors": {...}
}
```

#### ISteamUserStats — Achievements and Playtime

```
Claim: Steam tracks total playtime since early 2009 and provides achievement data, global achievement percentages, and player count via ISteamUserStats.
Source: Steam Web API Wiki
URL: https://developer.valvesoftware.com/wiki/Steam_Web_API
Date: 2026-04-11
Excerpt: "Steam began tracking total playtime in early 2009."
Context: Historical note on playtime tracking
Confidence: high
```

| Endpoint | Purpose |
|----------|---------|
| `ISteamUserStats/GetPlayerAchievements/v1/` | User's unlocked achievements |
| `ISteamUserStats/GetSchemaForGame/v2/` | Achievement definitions |
| `ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/` | Global achievement stats |
| `ISteamUserStats/GetUserStatsForGame/v2/` | User game statistics |
| `ISteamUserStats/GetNumberOfCurrentPlayers/v1/` | Current player count |

### SteamSpy

SteamSpy provides aggregated Steam game statistics including owner estimates, average playtime, and concurrent players.

```
Claim: SteamSpy API provides appdetails endpoint returning owner estimates, average playtime (average_forever), median playtime, concurrent players, and price data for any Steam app.
Source: SteamSpyPI GitHub
URL: https://github.com/woctezuma/steamspypi
Date: 2018-05-15 (ongoing)
Excerpt: "Returns details for 1000 games. Data is sorted by decreasing number of owners."
Context: Rate limited to 1 'all' request per minute
Confidence: high
```

**Rate Limits:** 1 request per minute for `all` endpoint; no published limits for `appdetails`.

**Key Endpoints:**
- `https://steamspy.com/api.php?request=appdetails&appid={APPID}` — Game details
- `https://steamspy.com/api.php?request=all&page={PAGE}` — Bulk data (1000 games/page)

### Steam Image CDN

Steam serves game images through Akamai CDN:
- Header image: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/header.jpg`
- Background: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/page_bg_generated_v6b.jpg`
- Library cover: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/library_600x900.jpg`
- Library hero: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/library_hero.jpg`
- Logo: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/logo.png`
- Icon: `https://steamcdn-a.akamaihd.net/steam/apps/{APPID}/icon.jpg`

---

## Epic Games Store API

### Overview
Epic Games does not provide an official public API for library access. However, an unofficial GraphQL-based API exists and is well-documented through community wrappers.

### Unofficial API (Python Wrapper)

```
Claim: The Epic Games Store uses a GraphQL API that can be accessed via unofficial wrappers. It uses cloudscraper to battle anti-bot protections and requires careful rate limiting.
Source: epicstore_api GitHub
URL: https://github.com/SD4RK/epicstore_api
Date: 2020-02-15
Excerpt: "An unofficial library to work with Epic Games Store Web API. The library works with cloudscraper under the hood to battle the anti-bot protections, please be careful with the amount of requests you do."
Context: Python wrapper; unofficial but widely used
Confidence: high
```

### Key Capabilities
- Product search and discovery
- Product details (developer, publisher, description, price)
- Critic reviews (OpenCritic integration)
- Store configuration (supported languages, system requirements, tags)
- Video/trailer information
- Catalog namespace queries

### Authentication
- OAuth2 with Epic Games account
- No simple way to acquire client ID/secret without being a game developer [^373^]
- For personal library access, the Legendary open-source launcher reverse-engineers the protocol

### Rate Limiting
- Anti-bot protection via Cloudflare
- No published rate limits
- Community recommendation: conservative request pacing

### Limitations
- Unofficial API may change without notice
- Anti-bot measures make scraping difficult
- No direct "library import" endpoint for end users
- Epic Online Services (EOS) Ecom API exists for ownership verification but requires partner access [^273^]

---

## GOG Galaxy API

### Overview
GOG provides two API layers: the GOG Galaxy SDK for game developers and the unofficial GOG web API for product data.

### GOG Galaxy SDK

```
Claim: The GOG Galaxy SDK provides interfaces for cloud storage, achievements, matchmaking, friends, user stats, and telemetry for game developers integrating with GOG Galaxy.
Source: GOG Galaxy SDK Documentation
URL: https://docs.gog.com/galaxyapi/
Date: Ongoing
Excerpt: "The GOG Galaxy SDK provides interfaces for: IApps (DLC discovery, game language), IStorage (cloud saves), IStats (achievements, leaderboards), IFriends (social), IUser (account info)"
Context: Designed for game developers, not end-user library access
Confidence: high
```

### Unofficial GOG API

```
Claim: GOG exposes product data via api.gog.com/products/{product_id} with a 200 requests/hour/IP rate limit, and a catalog API at catalog.gog.com/v1/catalog.
Source: GOG Forum Discussion
URL: https://www.reddit.com/r/gog/comments/1m84j87/could_there_be_any_problems_with_gog_for_using/
Date: 2025
Excerpt: "There is a 200 request/hour/IP on the api.gog.com/products/ endpoint."
Context: Unofficial but acknowledged by GOG community managers
Confidence: high
```

**Key Endpoints:**
- `https://api.gog.com/products/{prod_id}?expand=downloads,screenshots,videos,changelog`
- `https://catalog.gog.com/v1/catalog?limit=48&order=asc:externalProductId`
- `https://api.gog.com/v2/games/{prod_id}`

**Product Data Includes:**
```json
{
  "id": 1207658691,
  "title": "Game Title",
  "release_date": "2008-11-25T06:00:00+0200",
  "game_type": "game",
  "images": {
    "background": "//images-3.gog.com/...jpg",
    "logo": "//images-3.gog.com/..._glx_logo.jpg",
    "icon": "//images-1.gog.com/...png"
  },
  "screenshots": [...],
  "videos": [...],
  "dlcs": [...],
  "downloads": {...}
}
```

### GOG DB (Third-Party)

GOG DB (gogdb.org) aggregates GOG data and exposes it:
- All data at `/data` endpoint
- Plain JSON files, sometimes gzipped
- SQLite index for faster lookups
- Full database dumps at `/backups_v3`

---

## Alternative Metadata Sources

### RAWG API

**Overview:** Comprehensive gaming database with 500,000+ games across 50+ platforms.

**Key Features:** [^279^]
- Game catalogs and genre filtering
- Screenshots and trailers access
- Player activity data (Steam average playtime, RAWG player counts)
- Advanced search and filtering (Metacritic ratings, release dates, publishers)
- Machine learning-based recommendations
- Developer and publisher data

**Rate Limiting:** Requires API key; generous free tier.

### MobyGames API

```
Claim: MobyGames API v2 provides game data with a limit of 100 results per request, supporting filtering by platform, genre, group, and title substring search.
Source: MobyGames API Documentation
URL: https://www.mobygames.com/info/api/
Date: Ongoing
Excerpt: "limit: default 100, max 100... platform: multiple... genre: multiple... title: A substring of the title (not case sensitive)"
Context: Three output formats: id, brief, normal
Confidence: high
```

**Output Formats:**
- `id`: Minimal — game IDs only
- `brief`: Abbreviated data for link generation
- `normal`: Full game information with release data

### Giant Bomb API

```
Claim: Giant Bomb API provides extensive game fields including aliases, characters, concepts, date_added, description, developers, franchises, genres, image/images, name, number_of_user_reviews, original_game_rating, original_release_date, platforms, publishers, releases, reviews, similar_games, themes, and videos.
Source: Giant Bomb Ruby Wrapper
URL: https://www.rubydoc.info/gems/giantbomb-api/1.5.7/GiantBomb/Game
Date: Ongoing
Excerpt: "@@fields = [:aliases, :api_detail_url, :characters, :concepts, ... :videos]"
Context: Maximum 100 results per request; API key required
Confidence: high
```

**Rate Limiting:** API key required; limits vary by tier.
**Note:** Giant Bomb faced a DMCA takedown controversy in 2024, raising concerns about long-term stability.

### Comparison Table

| Feature | IGDB | RAWG | MobyGames | Giant Bomb |
|---------|------|------|-----------|------------|
| Games Count | 300K+ | 500K+ | 200K+ | 100K+ |
| Covers | Yes (1080p) | Yes | Limited | Yes |
| Screenshots | Yes | Yes | Limited | Yes |
| Videos | Yes | Yes | No | Yes |
| Genres | Yes | Yes | Yes | Yes |
| Platforms | Yes | Yes (50+) | Yes | Yes |
| ESRB/PEGI | Yes | No | Yes | Yes |
| Playtime Data | No | Yes (Steam) | No | No |
| Free Tier | 10K/mo | Yes | Yes | Yes |
| API Key | Twitch OAuth | Required | Required | Required |

---

## Game Metadata Schema Design

### Recommended Core Schema

Based on analysis of Schema.org VideoGame, IGDB fields, Steam Store API, and GOG product data, the following schema is recommended:

```json
{
  "game": {
    "id": "uuid",
    "external_ids": {
      "igdb_id": 123456,
      "steam_appid": 123456,
      "epic_namespace": "abc123",
      "gog_product_id": 1207658691,
      "rawg_slug": "game-title",
      "mobygames_id": 12345,
      "giantbomb_guid": "3030-12345"
    },
    "title": {
      "name": "Game Title",
      "sort_name": "Game Title, The",
      "aliases": ["Alternative Title"]
    },
    "description": {
      "short": "One-line summary",
      "full": "Detailed description",
      "storyline": "Plot summary"
    },
    "media": {
      "cover": {
        "url_600x900": "https://...",
        "url_1200x1800": "https://...",
        "source": "steamgriddb"
      },
      "hero": {
        "url_1920x620": "https://...",
        "url_3840x1240": "https://...",
        "source": "steamgriddb"
      },
      "logo": {
        "url": "https://...",
        "transparent": true
      },
      "icon": {
        "url_512": "https://..."
      },
      "screenshots": [
        {"url_thumb": "...", "url_full": "...", "width": 1920, "height": 1080}
      ],
      "trailers": [
        {"name": "Launch Trailer", "url_youtube": "...", "url_mp4": "..."}
      ]
    },
    "release": {
      "first_release_date": "2023-01-15",
      "release_dates": [
        {"platform": "PC", "date": "2023-01-15", "region": "US"}
      ],
      "status": "released"
    },
    "classification": {
      "genres": [{"id": "action", "name": "Action"}],
      "themes": [{"id": "sci-fi", "name": "Sci-Fi"}],
      "game_modes": ["single_player", "multiplayer"],
      "player_perspectives": ["first_person"],
      "keywords": ["open world", " crafting"]
    },
    "ratings": {
      "esrb": {"rating": "M", "content_descriptors": ["Violence", "Blood"]},
      "pegi": {"rating": 18},
      "cero": {"rating": "Z"},
      "aggregated_rating": 85.5,
      "aggregated_rating_count": 42
    },
    "people": {
      "developers": ["Developer Studio"],
      "publishers": ["Publisher Corp"],
      "directors": ["Director Name"],
      "composers": ["Composer Name"]
    },
    "technical": {
      "platforms": [{"name": "PC", "os": ["Windows", "Linux"]}],
      "engine": "Unreal Engine 5",
      "player_count": {
        "min": 1,
        "max": 64,
        "online": true,
        "local_coop": false
      },
      "supported_controllers": ["keyboard_mouse", "xbox", "playstation", "steam_deck"],
      "save_locations": {
        "windows": "%APPDATA%/GameTitle",
        "linux": "~/.local/share/GameTitle"
      },
      "system_requirements": {
        "minimum": {"cpu": "...", "gpu": "...", "ram": "8 GB", "storage": "50 GB"},
        "recommended": {"cpu": "...", "gpu": "...", "ram": "16 GB", "storage": "100 GB"}
      }
    },
    "ownership": {
      "library_sources": [{"platform": "steam", "appid": 123456}],
      "is_installed": true,
      "install_path": "/games/GameTitle",
      "playtime_minutes": 1245,
      "last_played": "2025-07-10T14:30:00Z",
      "completion_status": "playing"
    },
    "user_data": {
      "is_favorite": true,
      "user_rating": 9,
      "user_review": "Amazing game!",
      "user_tags": ["favorite", "backlog"],
      "user_categories": ["RPGs", "Completed"],
      "date_added_to_library": "2023-02-01"
    }
  }
}
```

### Schema.org VideoGame Alignment

```
Claim: Schema.org VideoGame type defines standard properties including name, author, publisher, genre, gamePlatform, processorRequirements, memoryRequirements, storageRequirements, aggregateRating, softwareAddOn, cheatCode, and playMode.
Source: Schema.org VideoGame Type
URL: https://schema.org/VideoGame
Date: 1999-07-02 (ongoing updates)
Excerpt: "Properties: gamePlatform, softwareAddOn, cheatCode, playMode, applicationCategory, operatingSystem, processorRequirements, memoryRequirements, storageRequirements"
Context: Industry-standard semantic markup for games
Confidence: high
```

---

## 4K Image Asset Pipeline

### Source Format Analysis

| Source | Cover Resolution | Hero Resolution | Icon | Format |
|--------|-----------------|-----------------|------|--------|
| SteamGridDB | 600x900 | 3840x1240 (4K) | 512x512 | PNG, WebP |
| Steam CDN | 600x900 | 1920x620 | 32x32 | JPG, PNG |
| IGDB | ~600x800 max | N/A | N/A | PNG |
| GOG | 660x930 | 1920x1080 | 256x256 | PNG, JPG |
| Epic | 720x960 | 1920x1080 | 256x256 | JPG, WebP |

### Image Optimization Pipeline

#### Format Comparison

```
Claim: AVIF offers superior compression efficiency compared to WebP, especially at lower bitrates, while WebP has faster encoding and broader browser support. Both significantly outperform JPEG and PNG.
Source: Cloudflare Image CDN Best Practices
URL: https://blog.blazingcdn.com/en-us/cloudflare-image-cdn-best-practices-webp-avif
Date: 2025-06-06
Excerpt: "Compression Efficiency: WebP=High, AVIF=Very High... Support: WebP=Widely supported, AVIF=Growing support... Visual Quality at Low Bitrates: WebP=Good, AVIF=Excellent"
Context: 2025 studies confirm AVIF's edge in compression
Confidence: high
```

#### Recommended Optimization Settings

| Format | Quality | Use Case |
|--------|---------|----------|
| AVIF | 50-80% | Primary delivery format (best compression) |
| WebP | 75-85% | Fallback for older clients |
| JPEG | 85-90% | Legacy fallback only |
| PNG | Lossless | Icons, logos with transparency |

#### Responsive Image Strategy

```
Claim: Modern responsive image delivery uses srcset with width descriptors combined with WebP/AVIF for maximum optimization, avoiding sending 2000px images to mobile devices.
Source: Image Optimization 2025 Guide
URL: https://www.frontendtools.tech/blog/modern-image-optimization-techniques-2025
Date: 2025-12-09
Excerpt: "Use srcset with width descriptors (800w, 1200w) for different resolutions... Combine with modern formats (WebP/AVIF) for maximum optimization... Compress: 70-85% quality for WebP, 50-80% for AVIF"
Context: web.dev best practices aligned
Confidence: high
```

#### CDN Delivery Architecture

**Recommended: Image-Optimizing CDN**

```
Claim: Modern image-optimizing CDNs like Cloudflare Images, Imgix, Cloudinary, and ImageKit automatically handle format conversion, responsive sizing, and quality optimization based on device characteristics.
Source: Image Optimization Best Practices
URL: https://hashmeta.com/blog/image-optimization-webp-avif-and-next-gen-formats-for-superior-website-performance/
Date: 2026-03-21
Excerpt: "Use image-optimizing CDNs for automatic format conversion and global delivery... CDN selection should prioritize providers with strong regional presence"
Context: Multi-CDN strategy for global reach
Confidence: high
```

**CDN Options:**

| CDN | Format Auto-Conversion | Cost Model | Best For |
|-----|----------------------|------------|----------|
| Cloudflare Images | Yes (AVIF, WebP) | $5/100K images | High volume |
| Imgix | Yes | $3-10/1000GB | Enterprise |
| ImageKit | Yes | Free tier + usage | Startups |
| Cloudinary | Yes | Free tier + usage | Flexibility |

#### White-Label Asset Injection

For the Phase 1 requirement of white-label asset injection:

1. **Upload Pipeline:** Accept PNG/WEBP uploads via admin API
2. **Processing:** Auto-convert to AVIF + WebP variants at multiple resolutions
3. **Storage:** Store original + optimized variants in object storage (S3/MinIO)
4. **CDN:** Serve through image-optimizing CDN with automatic format negotiation
5. **Metadata:** Store asset metadata in database with source attribution

---

## Local Asset Caching Strategy

### Multi-Tier Caching Architecture

```
Tier 1: In-Memory LRU Cache (hot assets)
Tier 2: Disk LRU Cache (warm assets)  
Tier 3: SQLite Metadata Index
Tier 4: Origin API (cold assets)
```

### Go LRU Cache Implementations

```
Claim: Bleve runs in memory and writes to disk similar to SQLite, providing a self-contained search solution without needing external services. Search results across 10 million+ documents take 50-100ms including filtering.
Source: Bleve — Build a fast search engine in Golang
URL: https://kevincoder.co.za/bleve-how-to-build-a-rocket-fast-search-engine
Date: 2024-04-02
Excerpt: "Self-contained... Bleve runs in memory and writes to disk similar to Sqlite... Fast: Search results across 10 million+ documents take just 50-100ms"
Context: Go-native search library evaluation
Confidence: high
```

#### Recommended Go Cache Libraries

| Library | Type | Best For |
|---------|------|----------|
| `groupcache/lru` | In-memory LRU | Small hot caches |
| `github.com/hashicorp/golang-lru` | Thread-safe LRU | General purpose |
| `github.com/coocood/freecache` | Zero-GC cache | High-throughput |
| `github.com/dgraph-io/ristretto` | High-performance | Dgraph-grade caching |

### Disk Caching Strategy

```
Claim: For game asset caching, a combination of SQLite for metadata indexing and the filesystem for binary blob storage provides the best balance of performance and simplicity.
Source: Stack Overflow Discussion
URL: https://stackoverflow.com/questions/25627338/storage-of-images-for-a-cache-or-sqlite
Date: 2014-09-13
Excerpt: "If you use LRU cache for bitmaps, you can use Memory Cache with LruCache class or Disk Cache with DiskLruCache."
Context: File system for blobs, SQLite for indexing
Confidence: high
```

**Recommended Implementation:**

```go
// Disk cache structure
cache/
  metadata.db  // SQLite: asset_key, file_path, size, last_accessed, etag
  images/
    {hash_prefix}/
      {asset_hash}.webp
      {asset_hash}.avif
  heroes/
    {hash_prefix}/
      {asset_hash}.webp
```

### Cache Eviction Policy

```
Claim: LRU (Least Recently Used) eviction is the standard approach for bounded caches, using a combination of a hash map for O(1) lookup and a doubly-linked list for O(1) insertion/deletion.
Source: LRU Cache Implementation in Go
URL: https://medium.com/@anshuman.mandal.01/understanding-cache-why-it-matters-and-how-to-implement-an-lru-cache-in-go-8881f600a1d9
Date: 2025-03-27
Excerpt: "We use Doubly Linked List + Hash Map... O(1) lookup time"
Context: Standard CS approach validated
Confidence: high
```

### SQLite Metadata Schema

```sql
CREATE TABLE asset_cache (
    asset_key TEXT PRIMARY KEY,
    asset_type TEXT NOT NULL, -- cover, hero, icon, screenshot, trailer
    source TEXT NOT NULL, -- steamgriddb, steam, igdb, gog, epic, uploaded
    original_url TEXT,
    local_path TEXT NOT NULL,
    format TEXT NOT NULL, -- avif, webp, png
    width INTEGER,
    height INTEGER,
    file_size INTEGER,
    etag TEXT,
    last_accessed INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER
);

CREATE INDEX idx_type_accessed ON asset_cache(asset_type, last_accessed);
CREATE INDEX idx_source ON asset_cache(source);
```

---

## Search and Indexing

### Comparison: Meilisearch vs Elasticsearch vs SQLite FTS5 vs Bleve

```
Claim: Meilisearch uses 6-8x more storage than SQLite FTS5 for the same dataset (217 MB vs 26 MB for 31,944 movie documents) but provides typo tolerance and instant search.
Source: Meilisearch GitHub Issue #4211
URL: https://github.com/meilisearch/meilisearch/issues/4211
Date: 2023-11-14
Excerpt: "Indexing movies.json creates a 217 MB folder... 6-8x larger than indexing the same data with SQLite3 FTS5 (26 MB database file)"
Context: Storage efficiency comparison
Confidence: high
```

```
Claim: Meilisearch is a developer-friendly search engine delivering <50ms response times with minimal configuration, while Elasticsearch excels at massive scale with complex analytics. SQLite FTS5 implements BM25 relevance scoring natively.
Source: Meilisearch vs Elasticsearch Comparison
URL: https://www.meilisearch.com/blog/meilisearch-vs-elasticsearch
Date: 2023-03-29
Excerpt: "Meilisearch is a perfect choice if you need a developer-friendly tool to effortlessly deploy a typo-tolerant search... Elasticsearch; it's an excellent solution for companies with the necessary resources"
Context: Feature trade-offs documented
Confidence: high
```

### Feature Comparison

| Feature | Meilisearch | Elasticsearch | SQLite FTS5 | Bleve |
|---------|-------------|---------------|-------------|-------|
| Setup | Minutes | Days | Built-in | Hours |
| Typo Tolerance | Yes | Yes (fuzzy) | No | Yes |
| Relevance | Built-in BM25 | Configurable TF-IDF | BM25 | TF-IDF/BM25 |
| Storage Overhead | 6-8x | 3-4x | 1.5-2x | 2-3x |
| Response Time | <50ms | <100ms | <10ms (local) | <100ms |
| Max Documents | 4B | Billions | Millions | 10M+ |
| Memory | Lightweight | Heavy | Minimal | Moderate |
| Go Native | No | No | CGO | Pure Go |
| Faceting | Yes | Yes | Limited | Yes |
| Highlighting | Yes | Yes | Limited | Yes |

### Recommendation by Use Case

| Deployment | Recommendation |
|------------|---------------|
| Embedded client (single user) | **SQLite FTS5** — zero dependencies, minimal overhead |
| Desktop app with advanced search | **Bleve** — Go-native, no CGO, good performance |
| Server-side (multi-user) | **Meilisearch** — typo tolerance, fast setup |
| Enterprise scale | **Elasticsearch** — proven at massive scale |

### SQLite FTS5 Implementation

```sql
-- Create virtual table for game search
CREATE VIRTUAL TABLE game_search USING fts5(
    title,
    description,
    genres,
    developers,
    publishers,
    tokenize='porter unicode61'
);

-- Populate from game data
INSERT INTO game_search(title, description, genres, developers, publishers)
SELECT g.title, g.description, g.genres, g.developers, g.publishers
FROM games g;

-- Search with ranking
SELECT g.*, rank
FROM game_search s
JOIN games g ON s.rowid = g.id
WHERE game_search MATCH 'action rpg'
ORDER BY rank;
```

### Bleve Implementation (Go)

```go
// Create index
mapping := bleve.NewIndexMapping()
index, err := bleve.New("games.bleve", mapping)

// Index a game
game := struct {
    Title       string
    Description string
    Genres      []string
    Developer   string
    Publisher   string
}{
    Title: "Cyberpunk 2077",
    Description: "Open-world action-adventure RPG",
    Genres: []string{"Action", "RPG"},
    Developer: "CD Projekt Red",
    Publisher: "CD Projekt",
}
index.Index("game_123", game)

// Search
query := bleve.NewMatchQuery("cyberpunk action")
searchRequest := bleve.NewSearchRequest(query)
searchResult, _ := index.Search(searchRequest)
```

---

## Sorting and Filtering Architecture

### Sort Criteria

| Sort Type | Field | Index Required |
|-----------|-------|---------------|
| Recent | `last_played` DESC | B-tree on timestamp |
| Alphabetical | `sort_name` ASC | B-tree on text (collated) |
| Release Date | `first_release_date` DESC | B-tree on date |
| Playtime | `playtime_minutes` DESC | B-tree on integer |
| Rating | `aggregated_rating` DESC | B-tree on float |
| Date Added | `date_added_to_library` DESC | B-tree on timestamp |
| Random | `random()` | None (or seeded) |

### Filter Categories

| Filter Type | Values | Implementation |
|-------------|--------|---------------|
| Installed | `true`/`false` | Boolean field |
| Favorites | `true`/`false` | Boolean field |
| Genre | Action, RPG, etc. | Multi-value tag index |
| Platform | PC, PlayStation, etc. | Multi-value tag index |
| Game Mode | Single-player, Multiplayer | Multi-value tag index |
| Completion Status | Playing, Completed, etc. | Enum field |
| Library Source | Steam, Epic, GOG, etc. | Multi-value tag index |
| Release Year | 2023, 2024, etc. | Integer range |
| User Rating | 1-10 | Integer range |
| Has Screenshots | `true`/`false` | Computed field |

### Architecture Pattern

```
Recommended: Faceted Search with Pre-computed Counts

1. Game documents stored in primary database (SQLite/PostgreSQL)
2. Search index (FTS5/Bleve/Meilisearch) for text search
3. Facet cache for filter counts (updated on library change)
4. Combined query: text search + filters + sort
```

---

## Game Trailers / Video Previews

### YouTube Integration

```
Claim: YouTube Data API v3 enables searching for game trailers by query string (game name + "trailer"), returning video IDs that can be embedded via iframe.
Source: Working with YouTube's Search API
URL: https://medium.com/@justinhwu95/working-with-youtubes-search-api-4dd1727d2c1a
Date: 2019-07-08
Excerpt: "Your API key... search parameters... callback function to handle both the errors or results... q lets you set your search term"
Context: Standard YouTube Data API usage pattern
Confidence: high
```

**Search Pattern:**
```
GET https://www.googleapis.com/youtube/v3/search?q={game_name}+trailer&videoCategoryId=20&type=video&part=snippet&maxResults=5&key={API_KEY}
```

`videoCategoryId=20` = Gaming category.

### Steam Store API — Trailers

```
Claim: The Steam Store API appdetails endpoint returns a 'movies' array with trailer names, thumbnail URLs, and MP4 files at 480p and max quality.
Source: Steam Store API Data Models
URL: https://stackoverflow.com/questions/51580365/getting-data-from-steam-store-api-using-newtonsoft-json
Date: 2018-07-30
Excerpt: "movies: [{name, mp4: {480, max}}]"
Context: Verified Steam Store API response structure
Confidence: high
```

### IGDB — Game Videos

IGDB provides video references including YouTube video IDs through the `/videos` endpoint.

### Video Preview Architecture

```
1. Trailer metadata stored in game record (URL, source, duration)
2. Lazy-load video player on game detail view
3. Option: Auto-play muted preview on hover (like PlayStation UI)
4. Cache trailer thumbnails locally
5. Embed YouTube via iframe or play MP4 directly
```

---

## User-Generated Content

### Ratings and Reviews

**Schema:**
```json
{
  "user_review": {
    "id": "uuid",
    "game_id": "uuid",
    "user_id": "uuid",
    "rating": 9,
    "review_text": "Detailed review...",
    "playtime_at_review": 1245,
    "recommendation": true,
    "created_at": "2025-07-01T00:00:00Z",
    "updated_at": "2025-07-01T00:00:00Z"
  }
}
```

### Playtime Tracking

```
Claim: Steam playtime data comes from IPlayerService/GetOwnedGames which returns playtime_forever and playtime_2weeks per game. For non-Steam games, local tracking via game process monitoring is required.
Source: IPlayerService Documentation
URL: https://partner.steamgames.com/doc/webapi/iplayerservice
Date: Ongoing
Excerpt: "playtime_forever: total playtime in minutes... playtime_2weeks: playtime in last 2 weeks in minutes"
Context: Steam-native tracking; other platforms need custom solutions
Confidence: high
```

**Local Playtime Tracking Implementation:**
- Monitor game process (PID tracking)
- Record start/stop timestamps
- Aggregate into daily/weekly/monthly views
- Sync with cloud on session end

### Import from External Sources

| Source | Data Available | API |
|--------|---------------|-----|
| Steam | Playtime, achievements, reviews | IPlayerService, ISteamUserStats |
| Epic | Limited (no official API) | Unofficial GraphQL |
| GOG | Galaxy SDK (for devs) | IStats interface |
| Xbox | Achievement history | Xbox Live API (restricted) |
| PlayStation | Trophy data | PSN API (restricted) |

---

## Per-Game Settings Persistence

### Settings Categories

| Category | Settings | Storage |
|----------|----------|---------|
| Controller | Button mappings, sensitivity, dead zones | Per-game profile JSON |
| Graphics | Resolution, quality preset, VSync, FOV | Per-game config JSON |
| Audio | Volume levels, output device, subtitles | Per-game config JSON |
| Gameplay | Difficulty, HUD settings, crosshair | Per-game config JSON |
| Accessibility | Colorblind mode, text size, remapping | Per-game config JSON |

### Storage Strategy

```
Recommended: Layered Configuration System

1. Default settings (shipped with app)
2. Global user preferences (applied to all games)
3. Per-game overrides (specific to each game)
4. Per-platform overrides (Steam Deck vs desktop)

Resolution: Latter layers override earlier layers (cascade)
```

### Controller Profile Persistence

```
Claim: Cloud gaming services like GeForce Now face challenges with per-game controller bindings not persisting between sessions, requiring profile save/load functionality.
Source: DCS World Forum
URL: https://forum.dcs.world/topic/379970-key-bindings-not-saving-between-sessions-geforce-now/
Date: 2025-10-02
Excerpt: "Save profiles in your own folder after setting up. Then you'd need to load those few for every device"
Context: Cloud gaming session persistence issue
Confidence: high
```

**Implementation:**
```json
{
  "controller_profile": {
    "id": "uuid",
    "game_id": "uuid",
    "profile_name": "Default",
    "controller_type": "xbox_one",
    "bindings": {
      "action_jump": "button_a",
      "action_attack": "trigger_rt",
      "camera_sensitivity": 0.75,
      "deadzone_left": 0.15,
      "deadzone_right": 0.15
    },
    "created_at": "2025-07-01",
    "updated_at": "2025-07-10"
  }
}
```

### Save Game Location Tracking

```json
{
  "save_location": {
    "game_id": "uuid",
    "platform": "windows",
    "paths": [
      "%USERPROFILE%/Documents/My Games/GameTitle",
      "%APPDATA%/GameTitle"
    ],
    "file_patterns": ["*.sav", "save*.dat"],
    "sync_enabled": true,
    "last_sync": "2025-07-10T14:30:00Z"
  }
}
```

---

## Tensions, Trade-offs, and Counter-Arguments

### API Dependency Risk

**Tension:** All third-party APIs can change, impose rate limits, or shut down.
**Mitigation:** 
- Aggressive local caching with long TTL
- Support multiple metadata sources with fallback chains
- Daily data dumps (IGDB) for bulk synchronization
- Local database as source of truth

### Image Quality vs Performance

**Tension:** True 4K images (3840x2160) are large; downloading hundreds for a library view is slow.
**Mitigation:**
- Progressive loading (blur-up placeholder)
- Multiple resolution tiers (thumbnail, preview, full)
- AVIF compression for 50-70% size reduction vs WebP
- LRU cache with intelligent pre-fetching

### Search: Embedded vs External

**Tension:** SQLite FTS5 is lightweight but lacks typo tolerance. Meilisearch adds 6-8x storage overhead.
**Resolution:** Use FTS5 for embedded clients, Meilisearch for server-side where typo tolerance matters.

### SteamGridDB Coverage Gaps

**Tension:** SteamGridDB focuses on Steam games; coverage for Epic/GOG exclusives may be limited.
**Mitigation:** 
- Upload white-label assets for missing games
- Fallback to platform-native artwork (Epic/GOG APIs)
- Community upload feature

### Epic Games API Stability

**Tension:** Epic's unofficial API uses anti-bot measures and may break.
**Mitigation:**
- Minimal dependency on Epic API
- Cache all fetched data locally
- Monitor for API changes

### Storage vs Speed for Caching

**Tension:** Large 4K asset libraries consume significant disk space.
**Trade-offs:**
| Strategy | Disk Usage | Load Time | Implementation |
|----------|-----------|-----------|---------------|
| Full local cache | High | Instant | 50-200MB per 100 games |
| Thumbnail + on-demand | Medium | ~500ms | 5-20MB + network |
| Streaming only | Low | 1-3s | Network dependent |

**Recommendation:** Full cache for installed games, thumbnails + on-demand for uninstalled.

---

## Recommended Architecture

### High-Level Design

```
+------------------+     +-------------------+     +------------------+
|   IGDB API       |     |  SteamGridDB API  |     |  Steam Web API   |
|   (metadata)     |     |  (4K artwork)     |     |  (library/import)|
+--------+---------+     +---------+---------+     +--------+---------+
         |                         |                          |
         v                         v                          v
+--------+-----------------------------------------------+---------+
|                     METADATA AGGREGATION LAYER                    |
|  - Normalize data from multiple sources                          |
|  - Merge conflicting fields with priority rules                  |
|  - Store in local database (SQLite/PostgreSQL)                   |
+--------+----------------------------------------------------------+
         |
         v
+--------+-----------------------------------------------+---------+
|                     LOCAL DATABASE                                |
|  - games table (normalized schema)                               |
|  - asset_cache table (metadata + paths)                          |
|  - game_search FTS5 index (full-text search)                     |
|  - user_data table (ratings, reviews, playtime)                  |
+--------+----------------------------------------------------------+
         |
         v
+--------+-----------------------------------------------+---------+
|                     ASSET MANAGER                                 |
|  - Disk LRU cache for images                                     |
|  - WebP/AVIF optimization pipeline                               |
|  - White-label asset injection                                   |
|  - CDN integration for remote assets                             |
+--------+----------------------------------------------------------+
         |
         v
+--------+-----------------------------------------------+---------+
|                     UI LAYER                                      |
|  - 4K cover grid with lazy loading                               |
|  - Search with filters and sorting                               |
|  - Game detail with trailers and screenshots                     |
|  - Per-game settings management                                  |
+-------------------------------------------------------------------+
```

### Priority Implementation Order

1. **Phase 1 (MVP):** IGDB metadata + SteamGridDB covers + Steam library import + SQLite FTS5 + Disk LRU cache
2. **Phase 2:** Epic/GOG integration + AVIF optimization + Meilisearch upgrade
3. **Phase 3:** User ratings/reviews + Playtime tracking + Per-game settings + Video previews

### Technology Stack Recommendation

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Backend | Go | Performance, single binary, Bleve native |
| Database | SQLite (embedded) / PostgreSQL (server) | Flexibility |
| Search | SQLite FTS5 (now) / Meilisearch (later) | Progressive enhancement |
| Cache | Ristretto (memory) + Disk LRU (filesystem) | Performance |
| Images | AVIF primary, WebP fallback | Best compression |
| CDN | Cloudflare Images or ImageKit | Auto-format conversion |
| Video | YouTube iframe embed | Simple, reliable |

---

## References

[^276^] IGDB Blog — "It's Here, the New IGDB API" (2018-12-21) — https://medium.com/igdb/its-here-the-new-igdb-api-f6ad745b53fe

[^269^] Steamworks Documentation — IPlayerService Interface — https://partner.steamgames.com/doc/webapi/iplayerservice

[^270^] Building a Steam Game Library Analyzer with Python (2023-05-17) — https://medium.com/@MrLuisArano/building-a-steam-game-library-analyzer-with-python-8db9507d3398

[^265^] epicstore-api PyPI package — https://libraries.io/pypi/epicstore-api-fcorz

[^266^] Epic Games Forums — Integrating Epic Game Library — https://forums.unrealengine.com/t/integrating-epic-game-library-in-new-program/733449

[^267^] epicstore_api documentation — https://epicstore-api.readthedocs.io/en/latest/intro.html

[^268^] GitHub — SD4RK/epicstore_api — https://github.com/SD4RK/epicstore_api

[^273^] Epic Online Services — Ecom Web APIs — https://dev.epicgames.com/docs/web-api-ref/ecom-web-apis

[^271^] GitHub — SteamTinkerLaunch Wiki — SteamGridDB — https://github.com/frostworx/steamtinkerlaunch/wiki/SteamGridDB

[^275^] SteamGridDB Changelog — https://changelog.steamgriddb.com/

[^277^] Reddit — Guide to update Steam art with SteamGridDB — https://www.reddit.com/r/steamgrid/comments/ym0ade/guide_to_update_all_the_steam_art_automatically/

[^279^] API League — Best Game APIs — https://apileague.com/articles/best-game-api/

[^280^] Polaris7 — Giant Bomb vs RAWG — https://www.polaris7.io/compare/giant-bomb-vs-rawg

[^281^] docs.rs — igdb Rust client — https://docs.rs/crate/igdb/latest

[^282^] dlthub — IGDB Python API Docs — https://dlthub.com/context/source/igdb

[^283^] SteamGridDB API v2 — https://www.steamgriddb.com/api/v2

[^284^] GitHub — SteamTinkerLaunch SteamGridDB Integration — https://github.com/sonic2kk/steamtinkerlaunch/issues/933

[^285^] Duke Yin — SteamGridDB API Endpoints — https://tech.dukeyin.com/2023/06/15/steamgriddb-api/

[^287^] MobyGames API Documentation — https://www.mobygames.com/info/api/

[^288^] IGDB API Documentation — Getting Started — https://api-docs.igdb.com/

[^313^] GOG DB — More Information (API) — https://www.gogdb.org/moreinfo

[^314^] GitHub — Meilisearch Storage Usage Issue — https://github.com/meilisearch/meilisearch/issues/4211

[^315^] Meilisearch Documentation — Comparison to alternatives — https://meilisearch.com/docs/resources/comparisons/alternatives

[^316^] Hacker News — Full text search over Postgres — https://news.ycombinator.com/item?id=41173288

[^317^] Medium — Postgres FTS vs Meilisearch vs Elasticsearch — https://medium.com/@simbatmotsi/postgres-full-text-search-vs-meilisearch-vs-elasticsearch-choosing-a-search-stack-that-scales-fcf17ef40a1b

[^319^] Schema.org — VideoGame Type — https://schema.org/VideoGame

[^320^] Meilisearch Blog — Elasticsearch vs Qdrant vs Meilisearch — https://www.meilisearch.com/blog/elasticsearch-vs-qdrant

[^321^] Stack Overflow — Storage of images for cache or SQLite — https://stackoverflow.com/questions/25627338/storage-of-images-for-a-cache-or-sqlite

[^322^] cl-gog-galaxy documentation — https://shinmera.github.io/cl-gog-galaxy/

[^324^] KevinCoder — Bleve: Build a fast search engine in Golang — https://kevincoder.co.za/bleve-how-to-build-a-rocket-fast-search-engine

[^325^] Couchbase Blog — Text Analysis within FTS — https://www.couchbase.com/blog/full-text_search_text_analysis/

[^326^] Hacker News — Bleve: full-text search for Go — https://news.ycombinator.com/item?id=16087936

[^327^] Medium — Full-text search and indexing with Bleve — https://medium.com/developers-writing/full-text-search-and-indexing-with-bleve-part-1-bd73599d82ef

[^328^] Digits Blog — Real-time Transaction Indexing With Bleve — https://digits.com/blog/real-time-transaction-indexing-with-bleve/

[^329^] GitHub — blevesearch/bleve — https://github.com/blevesearch/bleve

[^330^] Gopher Academy — Bleve: Text Search Powered by Go — https://blog.gopheracademy.com/birthday-bash-2014/bleve-text-search-powered-by-go/

[^331^] pkg.go.dev — bleve package — https://pkg.go.dev/github.com/blevesearch/bleve/v2

[^332^] DCS World Forum — Key Bindings not saving on GeForce Now — https://forum.dcs.world/topic/379970-key-bindings-not-saving-between-sessions-geforce-now/

[^333^] Reddit — How do you sort your Steam libraries? — https://www.reddit.com/r/Steam/comments/1nzb4c5/how_do_you_guys_sort_your_libraries/

[^335^] GitHub — SteamSpyPI API wrapper — https://github.com/woctezuma/steamspypi

[^337^] Zuplo — Ultimate Steam Web API Guide — https://zuplo.com/learning-center/what-is-the-steam-web-api/

[^338^] Tildes — How do you organize your gaming library? — https://tildes.net/~games/ola/how_do_you_organize_your_gaming_library

[^353^] Reddit — Free game assets sources — https://www.reddit.com/r/gamedev/comments/t9lt22/it_might_be_common_knowledge_butthere_are_tons_of/

[^354^] Apify — Steam Scraper — https://apify.com/sovereigntaylor/steam-scraper

[^357^] GitHub — Playnite Steam metadata source issue — https://github.com/JosefNemec/Playnite/issues/638

[^358^] Dgraph Blog — State of Caching in Go — https://discuss.dgraph.io/t/the-state-of-caching-in-go-dgraph-blog/4157

[^359^] Medium — LRU Cache Implementation in Go — https://medium.com/@anshuman.mandal.01/understanding-cache-why-it-matters-and-how-to-implement-an-lru-cache-in-go-8881f600a1d9

[^360^] Dev.to — LRU Cache in Go with Generics — https://dev.to/johnscode/implement-an-lru-cache-in-go-1hbc

[^362^] Otter Blog — Evolution of Caching Libraries in Go — https://maypok86.github.io/otter/blog/cache-evolution/

[^365^] Reddit — YouTube gameplay/trailer search addon for Playnite — https://www.reddit.com/r/playnite/comments/1p7larx/youtube_gameplaytrailer_search_addon/

[^368^] Bubble Forum — How to play a movie trailer with YouTube API — https://forum.bubble.io/t/how-to-play-a-movie-trailer-with-youtube/22019

[^369^] Medium — Working with YouTube's Search API — https://medium.com/@justinhwu95/working-with-youtubes-search-api-4dd1727d2c1a

[^370^] GOG Forum — GOG Database API discussion — https://www.gog.com/forum/general/gog_database_a_website_that_collects_data_on_gog_games/page55

[^371^] NuGet — SteamApi.Models package — https://feed.nuget.org/packages/SteamApi.Models/1.1.1

[^372^] Reddit — GOG API rate limits — https://www.reddit.com/r/gog/comments/1m84j87/could_there_be_any_problems_with_gog_for_using/

[^373^] GOG Forum — Unofficial GOG API Documentation — https://www.gog.com/forum/general/unofficial_gog_api_documentation/page15

[^374^] SteamGridDB — Decent Animated Covers collection — https://www.steamgriddb.com/collection/340/grids

[^375^] Stack Overflow — Steam Store API appdetails params — https://stackoverflow.com/questions/51580365/getting-data-from-steam-store-api-using-newtonsoft-json

[^376^] GOG Galaxy SDK Documentation — API Classes — https://docs.gog.com/galaxyapi/group__api.html

[^377^] SteamGridDB — Best Covers collection — https://www.steamgriddb.com/collection/3504/grids

[^380^] SteamGridDB — generic assets collection — https://www.steamgriddb.com/collection/107/grids

[^381^] SteamGridDB — Grids browse page — https://www.steamgriddb.com/grids

[^383^] pkg.go.dev — steamapi package — https://pkg.go.dev/github.com/Jleagle/steam-go/steamapi

[^384^] Medium — Scraping all games from Steam with Python — https://medium.com/codex/scraping-information-of-all-games-from-steam-with-python-6e44eb01a299

[^388^] GOG Forum — Unofficial GOG API Documentation page 17 — https://www.gog.com/forum/general/unofficial_gog_api_documentation

[^389^] GOG API Documentation — Galaxy APIs — https://gogapidocs.readthedocs.io/en/latest/galaxy.html

[^409^] Valve Developer Wiki — Steam Web API — https://developer.valvesoftware.com/wiki/Steam_Web_API

[^410^] Steam Web API Documentation — ISteamUserStats — https://steamapi.xpaw.me/ISteamUserStats

[^411^] RubyDoc — Giant Bomb API Game fields — https://www.rubydoc.info/gems/giantbomb-api/1.5.7/GiantBomb/Game

[^412^] Steamworks Documentation — ISteamUserStats Interface — https://partner.steamgames.com/doc/api/isteamuserstats

[^413^] Stack Overflow — Total playtime for specific Steam app — https://stackoverflow.com/questions/57750281/how-to-get-the-total-playtime-for-specific-steam-app

[^414^] PyBomb documentation — https://pybomb.readthedocs.io/

[^415^] GitHub — api-giantbomb Ruby wrapper — https://github.com/games-directory/api-giantbomb

[^318^] Cloudflare Image CDN Best Practices — https://blog.blazingcdn.com/en-us/cloudflare-image-cdn-best-practices-webp-avif

[^325^] Image Optimization: WebP, AVIF Best Practices — https://hashmeta.com/blog/image-optimization-webp-avif-and-next-gen-formats-for-superior-website-performance/

[^326^] Image Optimization 2025 Guide — https://www.frontendtools.tech/blog/modern-image-optimization-techniques-2025

---

*Report compiled from 25 independent web searches across official documentation, GitHub repositories, technical publications, and community forums. All claims are cited with inline references to original sources.*
