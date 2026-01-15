package v1

// PublicMethods defines API endpoints that don't require authentication.
// These are typically registration, login, and public content endpoints.
var PublicMethods = map[string]struct{}{
	"/api.v1.UserService/RegisterUser": {},
	"/api.v1.UserService/LoginUser":    {},
	"/api.v1.NoteService/ListNotes":    {},
	"/api.v1.NoteService/GetNote":      {},
	"/api.v1.CategoryService/ListCategories": {},
	"/api.v1.CategoryService/GetCategory":    {},
	"/api.v1.CategoryService/GetCategoryBySlug": {},
	"/api.v1.TagService/ListTags":      {},
	"/api.v1.TagService/GetTag":        {},
	"/api.v1.TagService/GetTagBySlug":  {},
	"/api.v1.AttachmentService/ListAttachments": {},
	// Note: CreateNote, UpdateNote, DeleteNote require authentication
}

// IsPublicMethod checks if a procedure path is public (no authentication required).
func IsPublicMethod(procedure string) bool {
	_, ok := PublicMethods[procedure]
	return ok
}

