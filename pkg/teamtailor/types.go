package teamtailor

import (
	teamtailorv1 "github.com/sebastienmelki/teamtailor-go-sdk-mcp/internal/gen/teamtailor/v1"
)

// The generated types live under internal/gen/, which Go forbids external
// modules from importing. These aliases re-export the request, response, and
// resource types under the public pkg/teamtailor path so that consumers in
// other modules can build requests and read responses directly.

// ListCandidatesRequest is the request shape for Client.ListCandidates.
type ListCandidatesRequest = teamtailorv1.ListCandidatesRequest

// ListCandidatesResponse is the response shape for Client.ListCandidates.
type ListCandidatesResponse = teamtailorv1.ListCandidatesResponse

// GetCandidateRequest is the request shape for Client.GetCandidate.
type GetCandidateRequest = teamtailorv1.GetCandidateRequest

// GetCandidateResponse is the response shape for Client.GetCandidate.
type GetCandidateResponse = teamtailorv1.GetCandidateResponse

// ListJobApplicationsRequest is the request shape for Client.ListJobApplications.
type ListJobApplicationsRequest = teamtailorv1.ListJobApplicationsRequest

// ListJobApplicationsResponse is the response shape for Client.ListJobApplications.
type ListJobApplicationsResponse = teamtailorv1.ListJobApplicationsResponse

// GetJobApplicationRequest is the request shape for Client.GetJobApplication.
type GetJobApplicationRequest = teamtailorv1.GetJobApplicationRequest

// GetJobApplicationResponse is the response shape for Client.GetJobApplication.
type GetJobApplicationResponse = teamtailorv1.GetJobApplicationResponse

// ListStagesRequest is the request shape for Client.ListStages.
type ListStagesRequest = teamtailorv1.ListStagesRequest

// ListStagesResponse is the response shape for Client.ListStages.
type ListStagesResponse = teamtailorv1.ListStagesResponse

// GetStageRequest is the request shape for Client.GetStage.
type GetStageRequest = teamtailorv1.GetStageRequest

// GetStageResponse is the response shape for Client.GetStage.
type GetStageResponse = teamtailorv1.GetStageResponse

// Candidate is the JSON:API resource of type `candidates`.
type Candidate = teamtailorv1.Candidate

// CandidateAttributes mirrors the JSON:API `attributes` object on a candidate.
type CandidateAttributes = teamtailorv1.CandidateAttributes

// CandidateRelationships captures the JSON:API `relationships` block on a candidate.
type CandidateRelationships = teamtailorv1.CandidateRelationships

// JobApplication is the JSON:API resource of type `job-applications`.
// It links a candidate to a job and carries the candidate's pipeline status.
type JobApplication = teamtailorv1.JobApplication

// JobApplicationAttributes mirrors the JSON:API `attributes` object on a job-application.
type JobApplicationAttributes = teamtailorv1.JobApplicationAttributes

// JobApplicationRelationships captures the JSON:API `relationships` block on a job-application.
type JobApplicationRelationships = teamtailorv1.JobApplicationRelationships

// Stage is the JSON:API resource of type `stages` — one column in a job's hiring pipeline.
type Stage = teamtailorv1.Stage

// StageAttributes mirrors the JSON:API `attributes` object on a stage.
type StageAttributes = teamtailorv1.StageAttributes

// StageRelationships captures the JSON:API `relationships` block on a stage.
type StageRelationships = teamtailorv1.StageRelationships

// Links is the JSON:API top-level / relationship `links` object.
type Links = teamtailorv1.Links

// Meta is the JSON:API top-level `meta` object returned alongside list responses.
type Meta = teamtailorv1.Meta

// ResourceIdentifier is the JSON:API `{ "type": "...", "id": "..." }` shape.
type ResourceIdentifier = teamtailorv1.ResourceIdentifier

// ToOneRelationship is a JSON:API to-one relationship.
type ToOneRelationship = teamtailorv1.ToOneRelationship

// ToManyRelationship is a JSON:API to-many relationship.
type ToManyRelationship = teamtailorv1.ToManyRelationship

// JsonApiError is one error object inside a JSON:API error response.
type JsonApiError = teamtailorv1.JsonApiError

// JsonApiErrorResponse is the top-level JSON:API error document.
type JsonApiErrorResponse = teamtailorv1.JsonApiErrorResponse

// ErrorSource pinpoints the cause of an error inside a request.
type ErrorSource = teamtailorv1.ErrorSource
