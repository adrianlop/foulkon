package api

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
	"github.com/tecsisa/authorizr/database"
)

// Group domain
type Group struct {
	ID       string    `json:"id, omitempty"`
	Name     string    `json:"name, omitempty"`
	Path     string    `json:"path, omitempty"`
	CreateAt time.Time `json:"createAt, omitempty"`
	Urn      string    `json:"urn, omitempty"`
	Org      string    `json:"org, omitempty"`
}

func (g Group) GetUrn() string {
	return g.Urn
}

// Group identifier to retrieve them from DB
type GroupIdentity struct {
	Org  string `json:"org, omitempty"`
	Name string `json:"name, omitempty"`
}

type GroupMembers struct {
	Users []User `json:"users, omitempty"`
}

// Add an Group to database if not exist
func (api AuthAPI) AddGroup(authenticatedUser AuthenticatedUser, org string, name string, path string) (*Group, error) {
	// Validate fields
	if !IsValidName(name) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", name),
		}
	}
	if !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidPath(path) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: path %v", path),
		}
	}

	group := createGroup(org, name, path)

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_CREATE_GROUP, []Group{group})
	if err != nil {
		return nil, err
	}
	if len(groupsFiltered) < 1 {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Check if group already exists
	_, err = api.GroupRepo.GetGroupByName(org, name)

	// Check if group could be retrieved
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		switch dbError.Code {
		// Group doesn't exist in DB, so we can create it
		case database.GROUP_NOT_FOUND:
			// Create group
			createdGroup, err := api.GroupRepo.AddGroup(group)

			// Check if there is an unexpected error in DB
			if err != nil {
				//Transform to DB error
				dbError := err.(*database.Error)
				return nil, &Error{
					Code:    UNKNOWN_API_ERROR,
					Message: dbError.Message,
				}
			}

			return createdGroup, nil
		default: // Unexpected error
			return nil, &Error{
				Code:    UNKNOWN_API_ERROR,
				Message: dbError.Message,
			}
		}
	} else {
		return nil, &Error{
			Code:    GROUP_ALREADY_EXIST,
			Message: fmt.Sprintf("Unable to create group, group with org %v and name %v already exists", org, name),
		}
	}

}

//  Add member to group
func (api AuthAPI) AddMember(authenticatedUser AuthenticatedUser, userID string, groupName string, org string) error {
	// Validate fields
	if !IsValidUserExternalID(userID) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: externalId %v", userID),
		}
	}
	if !IsValidOrg(org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidName(groupName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", groupName),
		}
	}

	// Call repo to retrieve the group
	groupDB, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, groupDB.Urn, GROUP_ACTION_ADD_MEMBER, []Group{*groupDB})
	if err != nil {
		return err
	}
	if len(groupsFiltered) < 1 {
		return &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, groupDB.Urn),
		}
	}

	// Call repo to retrieve the user
	userDB, err := api.GetUserByExternalId(authenticatedUser, userID)
	if err != nil {
		return err
	}

	// Call repo to retrieve the GroupUserRelation
	isMember, err := api.GroupRepo.IsMemberOfGroup(userDB.ID, groupDB.ID)
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	// Error handling
	if isMember {
		return &Error{
			Code:    USER_IS_ALREADY_A_MEMBER_OF_GROUP,
			Message: fmt.Sprintf("User: %v is already a member of Group: %v", userID, groupName),
		}
	}

	// Add Member
	err = api.GroupRepo.AddMember(userDB.ID, groupDB.ID)

	// Check if there is an unexpected error in DB
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return nil
}

//  Remove member from group
func (api AuthAPI) RemoveMember(authenticatedUser AuthenticatedUser, userID string, groupName string, org string) error {
	// Validate fields
	if !IsValidUserExternalID(userID) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: externalId %v", userID),
		}
	}
	if !IsValidOrg(org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidName(groupName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", groupName),
		}
	}

	// Call repo to retrieve the group
	groupDB, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, groupDB.Urn, GROUP_ACTION_REMOVE_MEMBER, []Group{*groupDB})
	if err != nil {
		return err
	}
	if len(groupsFiltered) < 1 {
		return &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, groupDB.Urn),
		}
	}

	// Call repo to retrieve the user
	userDB, err := api.GetUserByExternalId(authenticatedUser, userID)
	if err != nil {
		return err
	}

	// Call repo to check if user is a member of group
	isMember, err := api.GroupRepo.IsMemberOfGroup(userDB.ID, groupDB.ID)
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	if !isMember {
		return &Error{
			Code: USER_IS_NOT_A_MEMBER_OF_GROUP,
			Message: fmt.Sprintf("User with externalId %v is not a member of group with org %v and name %v",
				userDB.ExternalID, groupDB.Org, groupDB.Name),
		}
	}

	// Remove Member
	err = api.GroupRepo.RemoveMember(userDB.ID, groupDB.ID)

	// Check if there is an unexpected error in DB
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return nil
}

// List members of a group
func (api AuthAPI) ListMembers(authenticatedUser AuthenticatedUser, org string, groupName string) ([]string, error) {
	// Validate fields
	if !IsValidName(groupName) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", groupName),
		}
	}
	if !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}

	// Call repo to retrieve the group
	group, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return nil, err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_LIST_MEMBERS, []Group{*group})
	if err != nil {
		return nil, err
	}
	if len(groupsFiltered) < 1 {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Get Members
	members, err := api.GroupRepo.GetGroupMembers(group.ID)

	// Error handling
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return nil, &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	externalIDs := []string{}
	for _, m := range members {
		externalIDs = append(externalIDs, m.ExternalID)
	}

	return externalIDs, nil
}

// Remove group
func (api AuthAPI) RemoveGroup(authenticatedUser AuthenticatedUser, org string, name string) error {
	// Validate fields
	if !IsValidName(name) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", name),
		}
	}
	if !IsValidOrg(org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}

	// Call repo to retrieve the group
	group, err := api.GetGroupByName(authenticatedUser, org, name)
	if err != nil {
		return err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_DELETE_GROUP, []Group{*group})
	if err != nil {
		return err
	}
	if len(groupsFiltered) < 1 {
		return &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Remove group with given org and name
	err = api.GroupRepo.RemoveGroup(group.ID)

	// Error handling
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return nil
}

func (api AuthAPI) GetGroupByName(authenticatedUser AuthenticatedUser, org string, name string) (*Group, error) {
	// Validate fields
	if !IsValidName(name) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", name),
		}
	}
	if !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}

	// Call repo to retrieve the group
	group, err := api.GroupRepo.GetGroupByName(org, name)

	// Error handling
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		// Group doesn't exist in DB
		switch dbError.Code {
		case database.GROUP_NOT_FOUND:
			return nil, &Error{
				Code:    GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: dbError.Message,
			}
		default: // Unexpected error
			return nil, &Error{
				Code:    UNKNOWN_API_ERROR,
				Message: dbError.Message,
			}
		}
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_GET_GROUP, []Group{*group})
	if err != nil {
		return nil, err
	}

	// Check if we have our user authorized
	if len(groupsFiltered) > 0 {
		groupsFiltered := groupsFiltered[0]
		return &groupsFiltered, nil
	} else {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

}

func (api AuthAPI) GetGroupList(authenticatedUser AuthenticatedUser, org string, pathPrefix string) ([]GroupIdentity, error) {
	// Validate fields
	if len(org) > 0 && !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if len(pathPrefix) > 0 && !IsValidPath(pathPrefix) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: PathPrefix %v", pathPrefix),
		}
	}

	if len(pathPrefix) == 0 {
		pathPrefix = "/"
	}

	// Call repo to retrieve the groups
	groups, err := api.GroupRepo.GetGroupsFiltered(org, pathPrefix)

	// Error handling
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return nil, &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	// Check restrictions to list
	var urnPrefix string
	if len(org) == 0 {
		urnPrefix = "*"
	} else {
		urnPrefix = GetUrnPrefix(org, RESOURCE_GROUP, pathPrefix)
	}
	filteredGroups, err := api.GetAuthorizedGroups(authenticatedUser, urnPrefix, GROUP_ACTION_LIST_GROUPS, groups)
	if err != nil {
		return nil, err
	}

	// Transform to identifiers
	groupIDs := []GroupIdentity{}
	for _, g := range filteredGroups {
		groupIDs = append(groupIDs, GroupIdentity{
			Org:  g.Org,
			Name: g.Name,
		})
	}

	return groupIDs, nil
}

// Update Group to database if exist
func (api AuthAPI) UpdateGroup(authenticatedUser AuthenticatedUser, org string, groupName string, newName string, newPath string) (*Group, error) {
	// Validate fields
	if !IsValidName(newName) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", newName),
		}
	}
	if !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidPath(newPath) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: path %v", newPath),
		}
	}

	// Call repo to retrieve the group
	group, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return nil, err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_UPDATE_GROUP, []Group{*group})
	if err != nil {
		return nil, err
	}
	if len(groupsFiltered) < 1 {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Check if a group with "newName" already exists
	newGroup, err := api.GetGroupByName(authenticatedUser, org, newName)

	if err == nil && group.ID != newGroup.ID {
		// Group already exists
		return nil, &Error{
			Code:    GROUP_ALREADY_EXIST,
			Message: fmt.Sprintf("Group name: %v already exists", newName),
		}
	}

	if err != nil {
		if apiError := err.(*Error); apiError.Code == UNAUTHORIZED_RESOURCES_ERROR || apiError.Code == UNKNOWN_API_ERROR {
			return nil, err
		}
	}

	// Get Group updated
	groupToUpdate := createGroup(org, newName, newPath)

	// Check restrictions
	groupsFiltered, err = api.GetAuthorizedGroups(authenticatedUser, groupToUpdate.Urn, GROUP_ACTION_UPDATE_GROUP, []Group{groupToUpdate})
	if err != nil {
		return nil, err
	}
	if len(groupsFiltered) < 1 {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, groupToUpdate.Urn),
		}
	}

	// Update group
	group, err = api.GroupRepo.UpdateGroup(*group, newName, newPath, groupToUpdate.Urn)

	// Check unexpected DB error
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return nil, &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return group, nil

}

func (api AuthAPI) AttachPolicyToGroup(authenticatedUser AuthenticatedUser, org string, groupName string, policyName string) error {
	// Validate fields
	if !IsValidName(groupName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: Group name %v", groupName),
		}
	}
	if !IsValidOrg(org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidName(policyName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: Policy name %v", policyName),
		}
	}

	// Check if group exists
	group, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_ATTACH_GROUP_POLICY, []Group{*group})
	if err != nil {
		return err
	}
	if len(groupsFiltered) < 1 {
		return &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Check if policy exists
	policy, err := api.GetPolicyByName(authenticatedUser, org, policyName)
	if err != nil {
		return err
	}

	// Check existing relationship
	isAttached, err := api.GroupRepo.IsAttachedToGroup(group.ID, policy.ID)
	if err != nil {
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	if isAttached {
		// Unexpected error
		return &Error{
			Code:    POLICY_IS_ALREADY_ATTACHED_TO_GROUP,
			Message: fmt.Sprintf("Policy: %v is already attached to Group: %v", policy.Name, group.Name),
		}
	}

	// Attach Policy to Group
	err = api.GroupRepo.AttachPolicy(group.ID, policy.ID)

	if err != nil {
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return nil
}

func (api AuthAPI) DetachPolicyToGroup(authenticatedUser AuthenticatedUser, org string, groupName string, policyName string) error {
	// Validate fields
	if !IsValidName(groupName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: Group name %v", groupName),
		}
	}
	if !IsValidOrg(org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}
	if !IsValidName(policyName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: Policy name %v", policyName),
		}
	}

	// Check if group exists
	group, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_DETACH_GROUP_POLICY, []Group{*group})
	if err != nil {
		return err
	}
	if len(groupsFiltered) < 1 {
		return &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Check if policy exists
	policy, err := api.GetPolicyByName(authenticatedUser, org, policyName)
	if err != nil {
		return err
	}

	// Check existing relationship
	isAttached, err := api.GroupRepo.IsAttachedToGroup(group.ID, policy.ID)
	if err != nil {
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	if !isAttached {
		return &Error{
			Code: POLICY_IS_NOT_ATTACHED_TO_GROUP,
			Message: fmt.Sprintf("Policy with org %v and name %v is not attached to group with org %v and name %v",
				policy.Org, policy.Name, group.Org, group.Name),
		}

	}

	// Detach Policy to Group
	err = api.GroupRepo.DetachPolicy(group.ID, policy.ID)

	if err != nil {
		dbError := err.(*database.Error)
		return &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	return nil
}

func (api AuthAPI) ListAttachedGroupPolicies(authenticatedUser AuthenticatedUser, org string, groupName string) ([]string, error) {
	// Validate fields
	if !IsValidName(groupName) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: name %v", groupName),
		}
	}
	if !IsValidOrg(org) {
		return nil, &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", org),
		}
	}

	// Check if group exists
	group, err := api.GetGroupByName(authenticatedUser, org, groupName)
	if err != nil {
		return nil, err
	}

	// Check restrictions
	groupsFiltered, err := api.GetAuthorizedGroups(authenticatedUser, group.Urn, GROUP_ACTION_LIST_ATTACHED_GROUP_POLICIES, []Group{*group})
	if err != nil {
		return nil, err
	}
	if len(groupsFiltered) < 1 {
		return nil, &Error{
			Code: UNAUTHORIZED_RESOURCES_ERROR,
			Message: fmt.Sprintf("User with externalId %v is not allowed to access to resource %v",
				authenticatedUser.Identifier, group.Urn),
		}
	}

	// Call repo to retrieve the GroupPolicyRelations
	attachedPolicies, err := api.GroupRepo.GetAttachedPolicies(group.ID)

	// Error handling
	if err != nil {
		//Transform to DB error
		dbError := err.(*database.Error)
		return nil, &Error{
			Code:    UNKNOWN_API_ERROR,
			Message: dbError.Message,
		}
	}

	policyIDs := []string{}
	for _, p := range attachedPolicies {
		policyIDs = append(policyIDs, p.Name)
	}
	return policyIDs, nil
}

// Private helper methods

func createGroup(org string, name string, path string) Group {
	urn := CreateUrn(org, RESOURCE_GROUP, path, name)
	group := Group{
		ID:       uuid.NewV4().String(),
		Name:     name,
		Path:     path,
		CreateAt: time.Now().UTC(),
		Urn:      urn,
		Org:      org,
	}

	return group
}
