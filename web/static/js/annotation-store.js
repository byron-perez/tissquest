/**
 * AnnotationStore — in-memory session model for annotation editing.
 *
 * Invariants:
 *   - An annotoriousId appears in AT MOST ONE of _created, _updated, _deletedIds.
 *   - Editing a _created item updates it inside _created (never moves to _updated).
 *   - Deleting a _created item removes it from _created only (never added to _deletedIds).
 *   - Editing a _persisted item moves its new state into _updated.
 *   - Deleting a _persisted item removes it from _persisted/_updated and adds id to _deletedIds.
 *
 * Feature: annotation-session-management
 */
function createAnnotationStore(persistedAnnotations) {
  // Internal state
  const _persisted  = new Map(); // annotoriousId → annotationObject
  const _created    = new Map(); // annotoriousId → annotationObject (new this session)
  const _updated    = new Map(); // annotoriousId → annotationObject (edits to persisted)
  const _deletedIds = new Set(); // annotoriousIds scheduled for server-side deletion

  // Initialise persisted set from server load
  (persistedAnnotations || []).forEach(function(a) {
    _persisted.set(a.id, a);
  });

  return {
    /**
     * Record a newly drawn annotation as pending.
     * Requirement 1.2 — no network call.
     */
    add: function(annotation) {
      _created.set(annotation.id, annotation);
    },

    /**
     * Update an annotation that already exists in the store.
     * - If it was created this session → stays in _created.
     * - If it was persisted → moves into _updated.
     * Requirement 1.3 — no network call.
     */
    update: function(annotation) {
      if (_created.has(annotation.id)) {
        _created.set(annotation.id, annotation);
      } else if (_persisted.has(annotation.id)) {
        _updated.set(annotation.id, annotation);
      }
    },

    /**
     * Remove a Pending_Annotation (never persisted).
     * Does NOT add to _deletedIds.
     * Requirements 1.4, 2.4 — no network call.
     */
    deletePending: function(id) {
      _created.delete(id);
    },

    /**
     * Schedule a persisted annotation for deletion on next save.
     * Requirements 2.2 — no network call.
     */
    deletePersisted: function(id) {
      _persisted.delete(id);
      _updated.delete(id);
      _deletedIds.add(id);
    },

    /**
     * Returns true if the id belongs to a pending (never-saved) annotation.
     * Used by the delete button to route to the correct delete method.
     */
    isPending: function(id) {
      return _created.has(id);
    },

    /**
     * Returns the full Session_Diff accumulated since last commitDiff.
     */
    getDiff: function() {
      return {
        created:    Array.from(_created.values()),
        updated:    Array.from(_updated.values()),
        deletedIds: Array.from(_deletedIds),
      };
    },

    /**
     * Returns true when there are no pending changes.
     * Requirement 3.3.
     */
    isEmpty: function() {
      return _created.size === 0 && _updated.size === 0 && _deletedIds.size === 0;
    },

    /**
     * Returns all currently visible annotations:
     * persisted (excluding deletions) + overrides from _updated + new from _created.
     */
    getAll: function() {
      var result = [];
      _persisted.forEach(function(a, id) {
        if (!_deletedIds.has(id)) {
          result.push(_updated.has(id) ? _updated.get(id) : a);
        }
      });
      _created.forEach(function(a) {
        result.push(a);
      });
      return result;
    },

    /**
     * Called after a successful batch save.
     * Clears the diff and rebuilds _persisted from the server's authoritative list.
     * Requirement 3.4.
     */
    commitDiff: function(serverAnnotations) {
      _created.clear();
      _updated.clear();
      _deletedIds.clear();
      _persisted.clear();
      (serverAnnotations || []).forEach(function(a) {
        _persisted.set(a.id, a);
      });
    },
  };
}

// Attach to window for use in non-module scripts (slide_viewer.html)
if (typeof window !== 'undefined') {
  window.createAnnotationStore = createAnnotationStore;
}

// ES module export for test harness
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { createAnnotationStore };
}
