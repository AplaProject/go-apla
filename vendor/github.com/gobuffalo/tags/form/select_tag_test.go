package form_test

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/tags"
	"github.com/gobuffalo/tags/form"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func Test_SelectTag(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	s := f.SelectTag(tags.Options{})
	r.Equal(`<select></select>`, s.String())
}

func Test_SelectTagWithOptions(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	s := f.SelectTag(tags.Options{
		"options": []map[string]interface{}{
			{"1/2 day": 1},
			{"1-2 days": 2},
			{"1 week": 7},
			{"1-2 weeks": 14},
		},
	})
	r.Equal(`<select><option value="1">1/2 day</option><option value="2">1-2 days</option><option value="7">1 week</option><option value="14">1-2 weeks</option></select>`, s.String())
}

func Test_SelectTagWithOptionsSelected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	s := f.SelectTag(tags.Options{
		"options": []map[string]interface{}{
			{"1/2 day": 1},
			{"1-2 days": 2},
			{"1 week": 7},
			{"1-2 weeks": 14},
		},
		"value": 1,
	})
	r.Equal(`<select><option value="1" selected>1/2 day</option><option value="2">1-2 days</option><option value="7">1 week</option><option value="14">1-2 weeks</option></select>`, s.String())
}

func Test_SelectTag_WithSelectOptions(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": form.SelectOptions{
			{Value: 1, Label: "one"},
			{Value: 2, Label: "two"},
		},
	})
	s := st.String()
	r.Contains(s, `<option value="1">one</option>`)
	r.Contains(s, `<option value="2">two</option>`)
}

func Test_SelectTag_WithSelectOptions_Selected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": form.SelectOptions{
			{Value: 3, Label: "three"},
			{Value: 2, Label: "two"},
		},
		"value": "3",
	})
	s := st.String()
	r.Contains(s, `<option value="3" selected>three</option>`)
	r.Contains(s, `<option value="2">two</option>`)
}

func Test_SelectTag_WithMap(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": map[string]interface{}{
			"one": 1,
			"two": 2,
		},
	})
	s := st.String()
	r.Contains(s, `<option value="1">one</option>`)
	r.Contains(s, `<option value="2">two</option>`)
}

func Test_SelectTag_WithMap_Selected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": map[string]interface{}{
			"three": 3,
			"two":   2,
		},
		"value": 3,
	})
	s := st.String()
	r.Contains(s, `<option value="3" selected>three</option>`)
	r.Contains(s, `<option value="2">two</option>`)
}

func Test_SelectTag_WithMap_Selected_Withplaceholder(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"placeholder": "Select a country",
		"options": []map[string]interface{}{
			{"Colombia": "CO"},
			{"France": "FR"},
			{"United States": "US"},
		},
	})
	s := st.String()
	r.Contains(s, `<option value="" hidden disabled>Select a country</option>`)
	r.Contains(s, `<option value="CO">Colombia</option>`)
	r.Contains(s, `<option value="FR">France</option>`)
	r.Contains(s, `<option value="US">United States</option>`)
}

func Test_SelectTag_WithMap_Selected_Withoutplaceholder(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": []map[string]interface{}{
			{"Colombia": "CO"},
			{"France": "FR"},
			{"United States": "US"},
		},
	})
	s := st.String()
	r.NotContains(s, `<option value="" hidden disabled></option>`)
	r.Contains(s, `<option value="CO">Colombia</option>`)
	r.Contains(s, `<option value="FR">France</option>`)
	r.Contains(s, `<option value="US">United States</option>`)
}

func Test_SelectTag_WithMap_Selected_Withplaceholder_Selected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"placeholder": "Select a country",
		"options": []map[string]interface{}{
			{"Colombia": "CO"},
			{"France": "FR"},
			{"United States": "US"},
		},
		"value": "CO",
	})
	s := st.String()
	r.Contains(s, `<option value="" hidden disabled>Select a country</option>`)
	r.Contains(s, `<option value="CO" selected>Colombia</option>`)
	r.Contains(s, `<option value="FR">France</option>`)
	r.Contains(s, `<option value="US">United States</option>`)
}

func Test_SelectTag_WithSlice(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": []string{"one", "two"},
	})
	s := st.String()
	r.Contains(s, `<option value="one">one</option>`)
	r.Contains(s, `<option value="two">two</option>`)
}

func Test_SelectTag_WithSlice_Selected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": []string{"one", "two"},
		"value":   "two",
	})
	s := st.String()
	r.Contains(s, `<option value="one">one</option>`)
	r.Contains(s, `<option value="two" selected>two</option>`)
}

func Test_SelectTag_WithSlice_Selectable(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": []SelectableModel{
			{"John", "1"},
			{"Peter", "2"},
		},
		"value": "1",
	})
	s := st.String()
	r.Contains(s, `<option value="1" selected>John</option>`)
	r.Contains(s, `<option value="2">Peter</option>`)
}

func Test_SelectTag_WithSlice_Selectable_Interface(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"options": []SelectableModel{
			{"John", "1"},
			{"Peter", "2"},
		},
		"value": SelectableModel{"John", "1"},
	})
	s := st.String()
	r.Contains(s, `<option value="1" selected>John</option>`)
	r.Contains(s, `<option value="2">Peter</option>`)
}

func Test_SelectTag_WithUUID_Selected(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	jid, _ := uuid.NewV4()
	pid, _ := uuid.NewV4()
	st := f.SelectTag(tags.Options{
		"options": []SelectableUUIDModel{
			{"John", jid},
			{"Peter", pid},
		},
		"value": pid,
	})
	s := st.String()
	r.Contains(s, fmt.Sprintf(`<option value="%s">John</option>`, jid))
	r.Contains(s, fmt.Sprintf(`<option value="%s" selected>Peter</option>`, pid))
}

func Test_SelectTag_WithUUID_Selected_withBlank(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	jid, _ := uuid.NewV4()
	pid, _ := uuid.NewV4()
	st := f.SelectTag(tags.Options{
		"options": []SelectableUUIDModel{
			{"John", jid},
			{"Peter", pid},
		},
		"value":       pid,
		"allow_blank": true,
	})
	s := st.String()
	r.Contains(s, `<option value=""></option>`)
	r.Contains(s, fmt.Sprintf(`<option value="%s">John</option>`, jid))
	r.Contains(s, fmt.Sprintf(`<option value="%s" selected>Peter</option>`, pid))
}

func Test_SelectTag_WithUUID_Selected_withBlankSelectOptions(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	jid, _ := uuid.NewV4()
	pid, _ := uuid.NewV4()
	st := f.SelectTag(tags.Options{
		"options": form.SelectOptions{
			form.SelectOption{Label: "John", Value: jid},
			form.SelectOption{Label: "Peter", Value: pid},
		},
		"value":       pid,
		"allow_blank": true,
	})
	s := st.String()
	r.Contains(s, `<option value=""></option>`)
	r.Contains(s, fmt.Sprintf(`<option value="%s">John</option>`, jid))
	r.Contains(s, fmt.Sprintf(`<option value="%s" selected>Peter</option>`, pid))
}

func Test_SelectTag_WithUUID_Selected_withoutBlankSelectOptions(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	jid, _ := uuid.NewV4()
	pid, _ := uuid.NewV4()
	st := f.SelectTag(tags.Options{
		"options": form.SelectOptions{
			form.SelectOption{Label: "John", Value: jid},
			form.SelectOption{Label: "Peter", Value: pid},
		},
		"value":       pid,
		"allow_blank": false,
	})
	s := st.String()
	r.NotContains(s, `<option value=""></option>`)
	r.Contains(s, fmt.Sprintf(`<option value="%s">John</option>`, jid))
	r.Contains(s, fmt.Sprintf(`<option value="%s" selected>Peter</option>`, pid))
}

func Test_SelectTag_Multiple(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"multiple": true,
		"options":  []string{"one", "two"},
	})
	s := st.String()
	r.Contains(s, `<select multiple>`)
	r.Contains(s, `<option value="one">one</option>`)
	r.Contains(s, `<option value="two">two</option>`)
}

func Test_SelectTag_Multiple_SelectOne(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"multiple": true,
		"options":  []string{"one", "two"},
		"value":    "one",
	})
	s := st.String()
	r.Contains(s, `<select multiple>`)
	r.Contains(s, `<option value="one" selected>one</option>`)
	r.Contains(s, `<option value="two">two</option>`)
}

func Test_SelectTag_Multiple_SelectMultiple(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"multiple": true,
		"options":  []string{"one", "two", "three"},
		"value":    []string{"one", "two"},
	})
	s := st.String()
	r.Contains(s, `<select multiple>`)
	r.Contains(s, `<option value="one" selected>one</option>`)
	r.Contains(s, `<option value="two" selected>two</option>`)
	r.Contains(s, `<option value="three">three</option>`)
}

func Test_SelectTag_Multiple_SelectMultiple_Selectable_Interface(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"multiple": true,
		"options": []SelectableModel{
			{"John", "1"},
			{"Peter", "2"},
			{"Mark", "3"},
		},
		"value": []SelectableModel{
			{"John", "1"},
			{"Peter", "2"},
		},
	})
	s := st.String()
	r.Contains(s, `<select multiple>`)
	r.Contains(s, `<option value="1" selected>John</option>`)
	r.Contains(s, `<option value="2" selected>Peter</option>`)
	r.Contains(s, `<option value="3">Mark</option>`)
}


func Test_SelectTagWithHTML(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	s := f.SelectTag(tags.Options{
		"options": []map[string]interface{}{
			{"<b>Not Bold</b>": "<u>Not Underlined</u>"},
		},
	})
	r.Equal(`<select><option value="&lt;u&gt;Not Underlined&lt;/u&gt;">&lt;b&gt;Not Bold&lt;/b&gt;</option></select>`, s.String())
}

func Test_SelectTag_Multiple_SelectMultiple_SelectableMultiple_Interface(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	st := f.SelectTag(tags.Options{
		"multiple": true,
		"options": []SelectableMultipleModel{
			{"John", "1"},
			{"Peter", "2"},
			{"Mark", "3"},
		},
	})
	s := st.String()
	r.Contains(s, `<select multiple>`)
	r.Contains(s, `<option value="1">John</option>`)
	r.Contains(s, `<option value="2">Peter</option>`)
	r.Contains(s, `<option value="3" selected>Mark</option>`)
}

type SelectableModel struct {
	Name string
	ID   string
}

func (sm SelectableModel) SelectLabel() string {
	return sm.Name
}

func (sm SelectableModel) SelectValue() interface{} {
	return sm.ID
}

type SelectableMultipleModel struct {
	Name string
	ID   string
}

func (sm SelectableMultipleModel) SelectLabel() string {
	return sm.Name
}

func (sm SelectableMultipleModel) SelectValue() interface{} {
	return sm.ID
}

func (sm SelectableMultipleModel) IsSelected() bool {
	if sm.Name == "Mark" {
		return true
	}
	return false
}

type SelectableUUIDModel struct {
	Name string
	ID   uuid.UUID
}

func (sm SelectableUUIDModel) SelectLabel() string {
	return sm.Name
}

func (sm SelectableUUIDModel) SelectValue() interface{} {
	return sm.ID
}
