// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package api1

import (
	"bitbucket.org/enesyteam/papo-server/model"
	"net/http"
)

func (api *API) InitConversationNote() {
	api.BaseRoutes.Conversation.Handle("/notes", api.ApiSessionRequired(getConversationNotes)).Methods("GET")
	api.BaseRoutes.Conversation.Handle("/notes", api.ApiSessionRequired(createConversationNote)).Methods("POST")
	api.BaseRoutes.ConversationNote.Handle("/update", api.ApiSessionRequired(updateNote)).Methods("PUT")
}

func getConversationNotes(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		return
	}
	cId := c.Params.ConversationId
	if p, err := c.App.GetConversationNotes(cId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.ToJson()))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(model.ConversationNotesToJson(p)))
	}
}

func createConversationNote(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		return
	}

	note := model.ConversationNoteFromJson(r.Body)

	if note == nil {
		c.SetInvalidParam("body")
		return
	}

	cId := c.Params.ConversationId
	note.ConversationId = cId
	note.Creator = c.App.Session.UserId

	if len(note.Message) == 0 {
		c.SetInvalidParam("Nội dung ghi chú")
		return
	}

	var rNote *model.ConversationNote
	var err *model.AppError

	rNote, err = c.App.CreateConversationNote(note)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rNote.ToJson()))
}

// cần fix lại, không cho phép update note của người khác
// => check page_id
func updateNote(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireConversationId()
	if c.Err != nil {
		return
	}

	c.RequireNoteId()
	if c.Err != nil {
		return
	}

	noteId := c.Params.NoteId

	note := model.ConversationNoteFromJson(r.Body)
	if note == nil {
		c.SetInvalidParam("body")
		return
	}
	note.Id = noteId

	conversationId := c.Params.ConversationId
	note.ConversationId = conversationId
	note.UpdateAt = model.GetMillis()

	var rNote *model.ConversationNote
	var err *model.AppError

	rNote, err = c.App.UpdateConversationNote(note)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.ToJson()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rNote.ToJson()))
}