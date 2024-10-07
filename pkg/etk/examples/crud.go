package main

import (
	"slices"

	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func CRUD(c etk.Context) (etk.View, etk.React) {
	prefixView, prefixReact := c.Subcomp("prefix", comps.TextArea)

	personsVar := etk.State(c, "list/items", persons{
		{"Hans", "Emil"}, {"Max", "Mustermann"}, {"Roman", "Tisch"}})
	selectedVar := etk.BindState(c, "list/selected", 0)
	listView, listReact := c.Subcomp("list", comps.ListBox)

	nameView, nameReact := c.Subcomp("name", comps.TextArea)
	nameContent := etk.BindState(c, "name/buffer", comps.TextBuffer{}).Get().Content

	surnameView, surnameReact := c.Subcomp("surname", comps.TextArea)
	surnameContent := etk.BindState(c, "surname/buffer", comps.TextBuffer{}).Get().Content

	createFn := func() {
		personsVar.Swap(func(p persons) persons {
			return slices.Concat(p, persons{{nameContent, surnameContent}})
		})
	}
	createView, createReact := c.Subcomp("create",
		etk.WithInit(Button, "label", "Create", "submit", createFn))

	updateFn := func() {
		personsVar.Swap(func(p persons) persons {
			p = slices.Clone(p)
			p[selectedVar.Get()] = person{nameContent, surnameContent}
			return p
		})
	}
	updateView, updateReact := c.Subcomp("update",
		etk.WithInit(Button, "label", "Update", "submit", updateFn))

	deleteFn := func() {
		personsVar.Swap(func(p persons) persons {
			selected := selectedVar.Get()
			if selected == len(personsVar.Get())-1 {
				return p[:len(p)-1]
			}
			return slices.Concat(p[:selected], p[selected+1:])
		})
	}
	deleteView, deleteReact := c.Subcomp("delete",
		etk.WithInit(Button, "label", "Delete", "submit", deleteFn))

	return Form(c,
		FormComp{"Filter prefix: ", prefixView, prefixReact, false},
		FormComp{"", listView, listReact, false},
		FormComp{"Name:    ", nameView, nameReact, false},
		FormComp{"Surname: ", surnameView, surnameReact, false},
		FormComp{"", createView, createReact, nameContent == "" || surnameContent == ""},
		FormComp{"", updateView, updateReact, nameContent == "" || surnameContent == ""},
		FormComp{"", deleteView, deleteReact, len(personsVar.Get()) > 0},
	)
}

type persons []person

func (p persons) Len() int           { return len(p) }
func (p persons) Show(i int) ui.Text { return ui.T(p[i].surname + ", " + p[i].name) }

type person struct {
	name    string
	surname string
}
