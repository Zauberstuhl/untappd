package untappd_test

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/mdlayher/untappd"
)

// TestClientVenueInfoBadVenue verifies that Client.Venue.Info returns an error when
// an invalid venue is queried.
func TestClientVenueInfoBadVenue(t *testing.T) {
	venueID := -1
	sVenueID := strconv.Itoa(venueID)

	c, done := venueInfoTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		path := "/v4/venue/info/" + sVenueID + "/"
		if p := r.URL.Path; p != path {
			t.Fatalf("unexpected URL path: %q != %q", p, path)
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(invalidVenueErrJSON)
	})
	defer done()

	_, _, err := c.Venue.Info(venueID, false)
	assertInvalidVenueErr(t, err)
}

// TestClientVenueInfoCompactOK verifies that Client.Venue.Info properly requests compact
// venue output.
func TestClientVenueInfoCompactOK(t *testing.T) {
	c, done := venueInfoTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assertParameters(t, r, url.Values{
			"compact": []string{"true"},
		})

		// In the future, we may return compact canned venue data here.
		// For now, write a mostly empty JSON object is enough to get
		// test coverage.
		w.Write([]byte(`{"response":{"venue":{"id":1}}}`))
	})
	defer done()

	if _, _, err := c.Venue.Info(1, true); err != nil {
		t.Fatal(err)
	}
}

// TestClientVenueInfoOK verifies that Client.Venue.Info returns a valid venue when
// provided with correct input parameters.
func TestClientVenueInfoOK(t *testing.T) {
	venueID := 1021
	sVenueID := strconv.Itoa(venueID)

	c, done := venueInfoTestClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		path := "/v4/venue/info/" + sVenueID + "/"
		if p := r.URL.Path; p != path {
			t.Fatalf("unexpected URL path: %q != %q", p, path)
		}

		w.Write(venueJSON)
	})
	defer done()

	v, _, err := c.Venue.Info(venueID, false)
	if err != nil {
		t.Fatal(err)
	}

	if id := v.ID; id != venueID {
		t.Fatalf("unexpected ID: %d != %d", id, venueID)
	}
	venueName := "Bell's Eccentric Cafe & General Store"
	if n := v.Name; n != venueName {
		t.Fatalf("unexpected Name: %q != %q", n, venueName)
	}
	venueCity := "Kalamazoo"
	if c := v.Location.City; c != venueCity {
		t.Fatalf("unexpected Location.City: %q != %q", c, venueCity)
	}
	foursquareID := "4a8f8efcf964a520761520e3"
	if c := v.Foursquare.ID; c != foursquareID {
		t.Fatalf("unexpected Foursquare.ID: %q != %q", c, foursquareID)
	}
	foursquareURL := "http://4sq.com/dheQpl"
	if c := v.Foursquare.URL; c != foursquareURL {
		t.Fatalf("unexpected Foursquare.URL: %q != %q", c, foursquareURL)
	}

	beerName := "Beer Name"
	if c := v.TopBeers[0].Name; c != beerName {
		t.Fatalf("unexpected TopBeers[0].Name: %q != %q", c, beerName)
	}
	beerBrewery := "Brewery Name"
	if c := v.TopBeers[0].Brewery.Name; c != beerBrewery {
		t.Fatalf("unexpected TopBeers[0].Brewery.Name: %q != %q", c, beerBrewery)
	}
	if c := v.Checkins[0].Beer.Name; c != beerName {
		t.Fatalf("unexpected Checkins[0].Beer.Name: %q != %q", c, beerName)
	}
	if c := v.Checkins[0].Brewery.Name; c != beerBrewery {
		t.Fatalf("unexpected Checkins[0].Brewery.Name: %q != %q", c, beerBrewery)
	}
}

// venueInfoTestClient builds upon testClient, and adds additional sanity checks
// for tests which target the venue info API.
func venueInfoTestClient(t *testing.T, fn func(t *testing.T, w http.ResponseWriter, r *http.Request)) (*untappd.Client, func()) {
	return testClient(t, func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		// Always GET request
		method := "GET"
		if m := r.Method; m != method {
			t.Fatalf("unexpected HTTP method: %q != %q", m, method)
		}

		// Always uses specific path prefix
		prefix := "/v4/venue/info/"
		if p := r.URL.Path; !strings.HasPrefix(p, prefix) {
			t.Fatalf("unexpected HTTP path prefix: %q != %q", p, prefix)
		}

		// Guard against panics
		if fn != nil {
			fn(t, w, r)
		}
	})
}

// Canned JSON used in tests
var venueJSON = []byte(`
{
  "meta": {
    "code": 200,
    "response_time": {
      "time": 0,
      "measure": "seconds"
    }
  },
  "notifications": {},
  "response": {
    "venue": {
      "venue_id": 1021,
      "venue_name": "Bell's Eccentric Cafe & General Store",
      "location": {
        "venue_city": "Kalamazoo"
      },
      "foursquare": {
        "foursquare_id": "4a8f8efcf964a520761520e3",
        "foursquare_url": "http://4sq.com/dheQpl"
      },
      "top_beers": {
        "offset": 0,
        "limit": 15,
        "count": 1,
        "items": [
          {
            "created_at": "Mon, 02 May 2016 00:48:33 +0000",
            "total_count": 1,
            "your_count": 0,
            "beer": {
              "beer_name": "Beer Name"
            },
            "brewery": {
              "brewery_name": "Brewery Name"
            }
          }
        ]
      },
      "checkins": {
        "count": 1,
        "items": [
          {
            "created_at": "Sat, 21 May 2016 00:15:40 +0000",
            "beer": {
              "beer_name": "Beer Name"
            },
            "brewery": {
              "brewery_name": "Brewery Name"
            }
          }
        ]
      }
    }
  }
}`)
