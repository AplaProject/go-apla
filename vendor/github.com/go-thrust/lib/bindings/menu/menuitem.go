package menu

import . "github.com/go-thrust/lib/common"

type MenuItem struct {
	CommandID uint   `json:"command_id,omitempty"`
	Label     string `json:"label,omitempty"`
	GroupID   uint   `json:"group_id,omitempty"`
	SubMenu   *Menu  `json:"submenu,omitempty"`
	Type      string `json:"type,omitempty"`
	Checked   bool   `json:"checked"`
	Enabled   bool   `json:"enabled"`
	Visible   bool   `json:"visible"`
	Parent    *Menu  `json:"-"`
}

func NewMenuItem() *MenuItem {
	return &MenuItem{}
}

func (mi MenuItem) IsSubMenu() bool {
	return mi.SubMenu != nil
}

func (mi MenuItem) IsCheckItem() bool {
	return mi.Type == "check"
}

func (mi MenuItem) IsRadioItem() bool {
	return mi.Type == "radio"
}

func (mi MenuItem) IsGroupID(groupID uint) bool {
	return mi.GroupID == groupID
}

func (mi MenuItem) IsCommandID(commandID uint) bool {
	return mi.CommandID == commandID
}

func (mi MenuItem) HandleEvent() {
	Log.Print("EventType", mi.Type)
	switch mi.Type {
	case "check":
		Log.Print("Toggling Checked(", mi.Checked, ")", "to", "checked(", !mi.Checked, ")")
		mi.Parent.SetChecked(mi.CommandID, !mi.Checked)
	case "radio":
		Log.Print("Toggling RadioChecked(", mi.Checked, ")", "to", "checked(", !mi.Checked, ")")
		mi.Parent.ToggleRadio(mi.CommandID, mi.GroupID, !mi.Checked)
	}

}
