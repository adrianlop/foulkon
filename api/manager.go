package api

// User repository that contains all user operations for this domain
type UserRepo interface {
	// This method get a user with specified External ID.
	// If user exists, it will return the user with error param as nil
	// If user doesn't exists, it will return the error code database.USER_NOT_FOUND
	// If there is an error, it will return error param with associated error message
	// and error code database.INTERNAL_ERROR
	GetUserByExternalID(id string) (*User, error)
	GetUserByID(id string) (*User, error)

	// This method store a user.
	// If there are a problem inserting user it will return an database.Error error
	AddUser(user User) (*User, error)

	GetUsersFiltered(pathPrefix string) ([]User, error)
	GetGroupsByUserID(id string) ([]Group, error)
	RemoveUser(id string) error
}

// Group repository that contains all user operations for this domain
type GroupRepo interface {
	GetGroupById(id string) (*Group, error)

	GetGroupByName(org string, name string) (*Group, error)
	GetGroupUserRelation(userID string, groupID string) (*GroupMembers, error)
	GetGroupPolicyRelation(groupID string, policyID string) (*GroupPolicies, error)
	GetListGroups(org string) ([]Group, error)
	RemoveGroup(org string, name string) error

	AddGroup(group Group) (*Group, error)
	AddMember(user User, group Group) error
	AttachPolicy(group Group, policy Policy) error
}

// Policy repository that contains all user operations for this domain
type PolicyRepo interface {
	GetPolicyById(id string) (*Policy, error)
	GetPolicyByName(org string, name string) (*Policy, error)
	AddPolicy(policy Policy) (*Policy, error)
	GetPoliciesFiltered(org string, pathPrefix string) ([]Policy, error)
}
