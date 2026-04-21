package collections

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreCollection "mcba/tissquest/internal/core/collection"
	coreTR "mcba/tissquest/internal/core/tissuerecord"
)

var (
	builderTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/collection_builder.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
		"web/templates/includes/collection_sections_list.html",
		"web/templates/includes/collection_assignments_list.html",
		"web/templates/includes/tr_search_results.html",
		"web/templates/includes/collection_tr_modal.html",
	}
	sectionsListTemplateFiles = []string{
		"web/templates/includes/collection_sections_list.html",
		"web/templates/includes/collection_assignments_list.html",
		"web/templates/includes/tr_search_results.html",
		"web/templates/includes/collection_tr_modal.html",
	}
	assignmentsListTemplateFiles = []string{
		"web/templates/includes/collection_assignments_list.html",
		"web/templates/includes/tr_search_results.html",
		"web/templates/includes/collection_tr_modal.html",
	}
)

func parseCollectionID(c *gin.Context) (uint, bool) {
	v, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid collection ID")
		return 0, false
	}
	return uint(v), true
}

func parseSectionID(c *gin.Context) (uint, bool) {
	v, err := strconv.ParseUint(c.Param("sid"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid section ID")
		return 0, false
	}
	return uint(v), true
}

// BuilderPage renders the full collection builder screen.
func BuilderPage(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}

	col, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	shared.RenderPage(c, builderTemplateFiles, "content", gin.H{
		"Title":      col.Name + " — Builder",
		"Collection": col,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Collections", URL: "/collections"},
			{Label: col.Name},
		},
	})
}

// UpdateCollectionMetadata handles HTMX PUT to update collection metadata from the builder.
func UpdateCollectionMetadata(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}

	existing, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	existing.Name = c.PostForm("name")
	existing.Description = c.PostForm("description")
	existing.Goals = c.PostForm("goals")
	collType := c.PostForm("type")
	if collType == "" {
		collType = "atlas"
	}
	existing.Type = coreCollection.CollectionType(collType)
	existing.Authors = c.PostForm("authors")

	if err := collectionService().UpdateCollection(id, existing); err != nil {
		shared.RenderError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	shared.SetFlash(c, "Collection metadata saved")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// NewSectionForm renders an inline form for creating a new section or subsection.
func NewSectionForm(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	var parentID *uint
	if pidStr := c.Query("parent_id"); pidStr != "" {
		if pid, err := strconv.ParseUint(pidStr, 10, 32); err == nil {
			v := uint(pid)
			parentID = &v
		}
	}
	newSectionFormTemplateFiles := []string{
		"web/templates/includes/collection_section_form.html",
	}
	shared.RenderFragment(c, newSectionFormTemplateFiles, "collection-section-form", gin.H{
		"CollectionID": id,
		"ParentID":     parentID,
	})
}

// NewSectionFormCancel clears the inline section form.
func NewSectionFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateSection creates a new section and returns the updated sections list.
func CreateSection(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}

	name := c.PostForm("name")
	var parentID *uint
	if pidStr := c.PostForm("parent_id"); pidStr != "" {
		if pid, err := strconv.ParseUint(pidStr, 10, 32); err == nil {
			v := uint(pid)
			parentID = &v
		}
	}

	if _, err := collectionService().CreateSection(id, name, parentID); err != nil {
		shared.RenderError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	shared.SetFlash(c, "Section created")
	// Clear the inline form container via OOB, then render the updated sections list
	c.Header("Content-Type", "text/html")
	// Render sections list as primary response
	col, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to reload collection")
		return
	}
	shared.RenderFragment(c, sectionsListTemplateFiles, "collection-sections-list", gin.H{
		"Collection": col,
	})
	// OOB: clear the form container
	fmt.Fprintf(c.Writer, `<div id="section-form-container" hx-swap-oob="true"></div>`)
}

// UpdateSection renames a section and returns the updated sections list.
func UpdateSection(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	name := c.PostForm("name")
	if err := collectionService().RenameSection(sid, name); err != nil {
		shared.RenderError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	shared.SetFlash(c, "Section renamed")
	renderSectionsList(c, id)
}

// DeleteSection deletes a section and returns the updated sections list.
func DeleteSection(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	if err := collectionService().DeleteSection(sid); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete section")
		return
	}

	shared.SetFlash(c, "Section deleted")
	renderSectionsList(c, id)
}

// ReorderSections persists new section positions and returns the updated sections list.
func ReorderSections(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}

	positions := parsePositionsMap(c, "positions")
	if err := collectionService().ReorderSections(id, positions); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to reorder sections")
		return
	}

	renderSectionsList(c, id)
}

// CreateAssignment assigns a tissue record to a section.
func CreateAssignment(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	trIDStr := c.PostForm("tissue_record_id")
	trID, err := strconv.ParseUint(trIDStr, 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	if _, err := collectionService().AssignTissueRecord(sid, uint(trID)); err != nil {
		if err == coreCollection.ErrDuplicateAssignment {
			c.Status(http.StatusConflict)
			c.Header("Content-Type", "text/html")
			c.String(http.StatusConflict, `<div class="alert alert-info text-sm">This tissue record is already assigned to this section.</div>`)
			return
		}
		shared.RenderError(c, http.StatusInternalServerError, "Failed to assign tissue record")
		return
	}

	shared.SetFlash(c, "Tissue record assigned")
	renderAssignmentsList(c, id, sid)
}

// DeleteAssignment removes a section assignment.
func DeleteAssignment(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	aidStr := c.Param("aid")
	aid, err := strconv.ParseUint(aidStr, 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	if err := collectionService().RemoveAssignment(uint(aid)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to remove assignment")
		return
	}

	shared.SetFlash(c, "Assignment removed")
	renderAssignmentsList(c, id, sid)
}

// ReorderAssignments persists new assignment positions.
func ReorderAssignments(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	positions := parsePositionsMap(c, "positions")
	if err := collectionService().ReorderAssignments(sid, positions); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to reorder assignments")
		return
	}

	renderAssignmentsList(c, id, sid)
}

// ViewCollection renders the public collection view page.
func ViewCollection(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}

	col, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	viewTemplateFiles := []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/collection_view.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}

	shared.RenderPage(c, viewTemplateFiles, "content", gin.H{
		"Title":      col.Name,
		"Collection": col,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Collections", URL: "/collections"},
			{Label: col.Name},
		},
	})
}

// renderSectionsList reloads the collection and renders the sections list fragment.
func renderSectionsList(c *gin.Context, collectionID uint) {
	col, err := collectionService().GetCollection(collectionID)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to reload collection")
		return
	}
	shared.RenderFragment(c, sectionsListTemplateFiles, "collection-sections-list", gin.H{
		"Collection": col,
	})
}

// renderAssignmentsList reloads the collection and renders the assignments list for a section.
func renderAssignmentsList(c *gin.Context, collectionID, sectionID uint) {
	col, err := collectionService().GetCollection(collectionID)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to reload collection")
		return
	}
	sec := findSection(col, sectionID)
	shared.RenderFragment(c, assignmentsListTemplateFiles, "collection-assignments-list", gin.H{
		"Collection": col,
		"Section":    sec,
	})
}

// findSection finds a section by ID within a collection (including subsections).
func findSection(col *coreCollection.Collection, sectionID uint) *coreCollection.Section {
	for i := range col.Sections {
		if col.Sections[i].ID == sectionID {
			return &col.Sections[i]
		}
		for j := range col.Sections[i].Subsections {
			if col.Sections[i].Subsections[j].ID == sectionID {
				return &col.Sections[i].Subsections[j]
			}
		}
	}
	return nil
}

// parsePositionsMap parses form fields like positions[123]=1 into a map[uint]int.
func parsePositionsMap(c *gin.Context, prefix string) map[uint]int {
	result := make(map[uint]int)
	c.Request.ParseForm()
	for key, vals := range c.Request.PostForm {
		if len(key) > len(prefix)+2 && key[:len(prefix)] == prefix {
			idStr := key[len(prefix)+1 : len(key)-1]
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				if len(vals) > 0 {
					if pos, err := strconv.Atoi(vals[0]); err == nil {
						result[uint(id)] = pos
					}
				}
			}
		}
	}
	return result
}

// SearchTissueRecordsForSection handles GET /tissue_records/search for the builder.
func SearchTissueRecordsForSection(c *gin.Context) {
	q := c.Query("q")
	sectionIDStr := c.Query("section_id")
	collectionIDStr := c.Query("collection_id")

	var sectionID uint
	if v, err := strconv.ParseUint(sectionIDStr, 10, 32); err == nil {
		sectionID = uint(v)
	}
	var collectionID uint
	if v, err := strconv.ParseUint(collectionIDStr, 10, 32); err == nil {
		collectionID = uint(v)
	}

	results, err := collectionService().SearchTissueRecords(q)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Search failed")
		return
	}

	searchTemplateFiles := []string{
		"web/templates/includes/tr_search_results.html",
	}
	shared.RenderFragment(c, searchTemplateFiles, "tr-search-results", gin.H{
		"Results":      results,
		"Query":        q,
		"SectionID":    sectionID,
		"CollectionID": collectionID,
	})
}

// MoveSection handles up/down reorder for a single section.
func MoveSection(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	direction := c.PostForm("direction") // "up" or "down"

	col, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	siblings, _ := getSiblings(col, sid)
	if siblings == nil {
		renderSectionsList(c, id)
		return
	}

	idx := -1
	for i, s := range siblings {
		if s.ID == sid {
			idx = i
			break
		}
	}
	if idx < 0 {
		renderSectionsList(c, id)
		return
	}

	positions := make(map[uint]int)
	for _, s := range siblings {
		positions[s.ID] = s.Position
	}

	if direction == "up" && idx > 0 {
		positions[siblings[idx].ID], positions[siblings[idx-1].ID] = positions[siblings[idx-1].ID], positions[siblings[idx].ID]
	} else if direction == "down" && idx < len(siblings)-1 {
		positions[siblings[idx].ID], positions[siblings[idx+1].ID] = positions[siblings[idx+1].ID], positions[siblings[idx].ID]
	}

	collectionService().ReorderSections(id, positions)
	renderSectionsList(c, id)
}

// MoveAssignment handles up/down reorder for a single assignment.
func MoveAssignment(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	aidStr := c.Param("aid")
	aid, err := strconv.ParseUint(aidStr, 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid assignment ID")
		return
	}

	direction := c.PostForm("direction")

	col, err := collectionService().GetCollection(id)
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	sec := findSection(col, sid)
	if sec == nil {
		renderAssignmentsList(c, id, sid)
		return
	}

	assignments := sec.Assignments
	idx := -1
	for i, a := range assignments {
		if a.ID == uint(aid) {
			idx = i
			break
		}
	}
	if idx < 0 {
		renderAssignmentsList(c, id, sid)
		return
	}

	positions := make(map[uint]int)
	for _, a := range assignments {
		positions[a.ID] = a.Position
	}

	if direction == "up" && idx > 0 {
		positions[assignments[idx].ID], positions[assignments[idx-1].ID] = positions[assignments[idx-1].ID], positions[assignments[idx].ID]
	} else if direction == "down" && idx < len(assignments)-1 {
		positions[assignments[idx].ID], positions[assignments[idx+1].ID] = positions[assignments[idx+1].ID], positions[assignments[idx].ID]
	}

	collectionService().ReorderAssignments(sid, positions)
	renderAssignmentsList(c, id, sid)
}

// getSiblings returns the sibling sections for a given section ID.
func getSiblings(col *coreCollection.Collection, sectionID uint) ([]coreCollection.Section, *uint) {
	for _, s := range col.Sections {
		if s.ID == sectionID {
			return col.Sections, nil
		}
	}
	for _, s := range col.Sections {
		for _, sub := range s.Subsections {
			if sub.ID == sectionID {
				return s.Subsections, &s.ID
			}
		}
	}
	return nil, nil
}

// CreateTissueRecordAndAssign creates a new TR and assigns it to a section.
func CreateTissueRecordAndAssign(c *gin.Context) {
	id, ok := parseCollectionID(c)
	if !ok {
		return
	}
	sid, ok := parseSectionID(c)
	if !ok {
		return
	}

	name := c.PostForm("name")
	notes := c.PostForm("notes")
	taxonIDStr := c.PostForm("taxon_id")

	if name == "" {
		c.Status(http.StatusUnprocessableEntity)
		c.Header("Content-Type", "text/html")
		c.String(http.StatusUnprocessableEntity, `<div class="alert alert-error text-sm">Name is required</div>`)
		return
	}

	tr := &coreTR.TissueRecord{
		Name:  name,
		Notes: notes,
	}
	if taxonIDStr != "" {
		if tid, err := strconv.ParseUint(taxonIDStr, 10, 32); err == nil && tid > 0 {
			uid := uint(tid)
			tr.TaxonID = &uid
		}
	}

	if err := collectionService().CreateTissueRecordAndAssign(tr, sid); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to create tissue record")
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Tissue record \"%s\" created and assigned", name))
	renderAssignmentsList(c, id, sid)
}
