package renderer

import (
	"encoding/json"
	"time"
)

type NotificationContext struct {
	Events []*Event `json:"Events"`

	CompanyID   int       `json:"CompanyID,omitempty"`
	CompanyName string    `json:"CompanyName,omitempty"`
	Now         time.Time `json:"Now,omitempty"`
}

type Event struct {
	Type         string `json:"Type,omitempty"`
	Description  string `json:"Description,omitempty"`
	IsActive     bool   `json:"IsActive,omitempty"`
	StartTime    string `json:"StartTime,omitempty"`
	EndTime      string `json:"EndTime,omitempty"`
	CurrentState string `json:"CurrentState,omitempty"`
	PreviousState string `json:"PreviousState,omitempty"`
	Importance   int    `json:"Importance,omitempty"`

	Details EventDetails `json:"-"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	type Alias Event

	aux := struct {
		*Alias
		Details EventDetails `json:"Details"`
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	e.Details = aux.Details

	return nil
}

type EventDetail struct {
	Name  string      `json:"Name"`
	Label string      `json:"Label,omitempty"`
	Value interface{} `json:"Value"`
	Tag   string      `json:"Tag,omitempty"`
}

type EventDetails []*EventDetail

func (d EventDetails) ToMap() map[string]interface{} {
	out := make(map[string]interface{}, len(d))
	for _, it := range d {
		if it == nil || it.Name == "" {
			continue
		}
		out[it.Name] = it.Value
	}
	return out
}

func (d EventDetails) WithTag(tag string) EventDetails {
	out := make(EventDetails, 0)
	for _, it := range d {
		if it != nil && it.Tag == tag {
			out = append(out, it)
		}
	}
	return out
}

func (d EventDetails) General() EventDetails {
	out := make(EventDetails, 0)
	for _, it := range d {
		if it != nil && it.Tag == "" {
			out = append(out, it)
		}
	}
	return out
}

func (d EventDetails) WithNames(names ...string) EventDetails {
	want := map[string]struct{}{}
	for _, n := range names {
		want[n] = struct{}{}
	}

	out := make(EventDetails, 0)
	for _, it := range d {
		if it == nil {
			continue
		}

		if _, ok := want[it.Name]; ok {
			out = append(out, it)
		}
	}
	return out
}

func (d EventDetails) Has(name string) bool {
	for _, it := range d {
		if it != nil && it.Name == name {
			return true
		}
	}
	return false
}

func (d EventDetails) HasTag(tag string) bool {
	for _, it := range d {
		if it != nil && it.Tag == tag {
			return true
		}
	}
	return false
}

func (d EventDetails) Get(name string) *EventDetail {
	for _, it := range d {
		if it != nil && it.Name == name {
			return it
		}
	}
	return &EventDetail{Name: name, Label: name, Value: nil}
}

func (d EventDetails) GetValue(name string) interface{} {
	return d.Get(name).Value
}

func (it *EventDetail) LabelOrName() string {
	if it == nil {
		return ""
	}

	if it.Label != "" {
		return it.Label
	}

	return it.Name
}
