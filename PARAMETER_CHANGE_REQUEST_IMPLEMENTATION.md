# Parameter Change Request Implementation

## Overview
Implemented a review and approval workflow for parameter updates. When updating parameters, instead of directly modifying them, users create change requests that must be reviewed and approved.

## Key Features
- ✅ One pending change request per parameter (enforces sequential reviews)
- ✅ Create change request with proposed changes
- ✅ View pending change request for a parameter
- ✅ View all change requests (history) for a parameter
- ✅ Approve change request → applies changes to parameter
- ✅ Reject change request → changes status only
- ✅ No deletion - use status changes instead

## Database Schema

### Table: `parameter_change_requests`
- `id` - Primary key
- `parameter_id` - FK to parameters
- `requested_by_user_id` - FK to users (who created the request)
- `status` - ENUM: pending, approved, rejected, cancelled
- `description` - Optional text describing the change
- `change_data` - JSONB containing proposed changes
- `reviewed_by_user_id` - FK to users (who approved/rejected)
- `reviewed_at` - Timestamp of review
- `created_at`, `updated_at`

## API Endpoints

### Create Change Request
```
POST /api/v1/parameters/:id/change-requests
Auth: Required (JWT)
Body: {
  "description": "Why this change is needed",
  "name": "new_parameter_name",
  "dataType": "boolean",
  "parameterDescription": "Updated description",
  "defaultRolloutValue": true,
  "rules": [...]
}
Response: ParameterChangeRequestResponse
```

### Get Pending Change Request
```
GET /api/v1/parameters/:id/change-requests/pending
Auth: Required (JWT)
Response: ParameterChangeRequestResponse | null
```

### Get All Change Requests (History)
```
GET /api/v1/parameters/:id/change-requests
Auth: Required (JWT)
Response: ParameterChangeRequestResponse[]
```

### Get Change Request by ID
```
GET /api/v1/parameter-change-requests/:id
Auth: Required (JWT)
Response: ParameterChangeRequestResponse
```

### Approve Change Request
```
PATCH /api/v1/parameter-change-requests/:id/approve
Auth: Required (JWT)
Body: {
  "comment": "Optional approval comment"
}
Response: ParameterChangeRequestResponse
```

### Reject Change Request
```
PATCH /api/v1/parameter-change-requests/:id/reject
Auth: Required (JWT)
Body: {
  "comment": "Optional rejection comment"
}
Response: ParameterChangeRequestResponse
```

## Workflow

1. **User wants to update parameter**
   - Calls `POST /api/v1/parameters/:id/change-requests`
   - System checks if there's already a pending request → rejects if exists
   - Creates change request with status `pending`

2. **Check pending requests**
   - Call `GET /api/v1/parameters/:id/change-requests/pending`
   - Returns the pending request or null

3. **Reviewer approves**
   - Calls `PATCH /api/v1/parameter-change-requests/:id/approve`
   - System applies all changes from `change_data` to the parameter
   - Updates rules (deletes old, creates new)
   - Updates `raw_value` for SDK
   - Enqueues sync job
   - Updates status to `approved`

4. **Reviewer rejects**
   - Calls `PATCH /api/v1/parameter-change-requests/:id/reject`
   - Updates status to `rejected`
   - No changes applied to parameter

## Files Created/Modified

### New Files
- `api/migrations/20251015200000_create_parameter_change_requests_table.up.sql`
- `api/migrations/20251015200000_create_parameter_change_requests_table.down.sql`
- `api/internal/model/parameter_change_request.go`
- `api/internal/repository/parameter_change_request.go`
- `api/internal/service/parameter_change_request.go`
- `api/internal/handler/parameter_change_request.go`

### Modified Files
- `api/internal/dto/parameter.go` - Added change request DTOs
- `api/internal/repository/repository.go` - Added repository interface methods
- `api/internal/service/service.go` - Added service interface methods
- `api/internal/router/router.go` - Added routes and handlers

## Change Request Data Structure

The `change_data` JSONB field stores:
```json
{
  "name": "new_name",
  "description": "new description",
  "dataType": "boolean",
  "defaultRolloutValue": true,
  "rules": [
    {
      "name": "Rule 1",
      "description": "Description",
      "type": "segment",
      "rolloutValue": false,
      "segmentId": 5,
      "matchType": "match"
    }
  ]
}
```

## Validation Rules
- Only one pending change request per parameter
- Must approve or reject before creating new request
- All changes are atomic (transaction-based)
- Status transitions: pending → approved/rejected (no going back)

## Next Steps
To use this system, you need to:
1. Run migrations: `make migrate-up`
2. Test the endpoints
3. Update frontend to use the new workflow

