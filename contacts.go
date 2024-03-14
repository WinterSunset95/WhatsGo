package main

import (
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func listOfContacts(searchString string, fullList map[types.JID]types.ContactInfo, fullListOfGroups []*types.GroupInfo) (contacts map[types.JID]types.ContactInfo) {
	// We make a copy of the full list of contacts
	contactsList := make(map[types.JID]types.ContactInfo);
	for jid, contact := range fullList {
		contactsList[jid] = contact;
	}

	// We make a copy of the full list of groups
	groupsList := make([]*types.GroupInfo, len(fullListOfGroups));
	copy(groupsList, fullListOfGroups);

	// We add the groups to the contacts list
	for i := 0; i < len(groupsList); i++ {
		groupToContact := types.ContactInfo{
			FullName: groupsList[i].Name,
		}
		contactsList[groupsList[i].JID] = groupToContact;
	}

	// First we filter out the contacts which don't have the 'FullName' field set
	for jid, contact := range contactsList {
		if contact.FullName == "" {
			delete(contactsList, jid)
		}
	}

	// If we have a search string, we filter out the contacts which don't match the search string
	if searchString != "" {
		for jid, contact := range contactsList {
			if !strings.Contains(contact.FullName, searchString) && !strings.Contains(jid.String(), searchString) {
				delete(contactsList, jid)
			}
		}
	}

	return contactsList;
}
