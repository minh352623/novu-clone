package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"CONVERDA/internal/messaging/application/service"
	"CONVERDA/internal/messaging/controller/dto"
	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/pkg/cursor"

	"github.com/google/uuid"
)

type conversationServiceImpl struct {
	msgRepo    domainRepo.MessageRepository
	threadRepo domainRepo.ThreadRepository
	subRepo    domainRepo.SubscriberRepository
	logRepo    domainRepo.AssignmentLogRepository
}

func NewConversationService(
	msgRepo domainRepo.MessageRepository,
	threadRepo domainRepo.ThreadRepository,
	subRepo domainRepo.SubscriberRepository,
	logRepo domainRepo.AssignmentLogRepository,
) service.ConversationService {
	return &conversationServiceImpl{
		msgRepo:    msgRepo,
		threadRepo: threadRepo,
		subRepo:    subRepo,
		logRepo:    logRepo,
	}
}

func (s *conversationServiceImpl) ReceiveMessage(ctx context.Context, tenantID, envID uuid.UUID, subKey, channel string, content []byte, senderID *uuid.UUID) (*entity.Message, error) {
	// 1. Find or Create Subscriber
	sub, err := s.subRepo.GetByKey(ctx, envID, subKey)
	if err != nil {
		sub = entity.NewSubscriber(envID, subKey)
		sub, err = s.subRepo.Create(ctx, sub)
		if err != nil {
			return nil, err
		}
	}

	// 2. Find Open Support Thread for this subscriber
	supportType := "support"
	threads, _, err := s.threadRepo.List(ctx, domainRepo.ThreadFilter{
		EnvironmentID: &envID,
		Type:          &supportType,
		MemberID:      &sub.ID, // Using MemberID field in filter to filter by participant
		Status:        nil,     // Any status except maybe resolved? Let's say we find any open one first.
	})

	var thread *entity.Thread
	for _, t := range threads {
		if t.Status != "resolved" {
			thread = t
			break
		}
	}

	if thread == nil {
		// Create new support thread
		if channel == "" {
			channel = "support" // Default channel if not provided
		}
		thread = entity.NewThread(envID, "support", channel)
		thread, err = s.threadRepo.Create(ctx, thread)
		if err != nil {
			return nil, err
		}
		// Add Subscriber as Participant
		part := entity.NewThreadParticipant(thread.ID, "subscriber", sub.ID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return nil, err
		}
		thread.Participants = append(thread.Participants, part)
	}

	// 3. Add Message
	realSenderID := sub.ID
	if senderID != nil {
		realSenderID = *senderID
	}

	msg := entity.NewMessage(tenantID, envID, thread.ID, "contact", &realSenderID, content)
	return s.msgRepo.Create(ctx, msg)
}

func (s *conversationServiceImpl) ReplyMessage(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, agentID uuid.UUID, content []byte) (*entity.Message, error) {
	// 1. Fetch Thread with environment check
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, fmt.Errorf("thread not found or environment mismatch: %w", err)
	}

	if thread.Status == "resolved" {
		return nil, errors.New("cannot reply to a resolved thread")
	}

	// Add Agent as Participant if not already
	isParticipant := false
	for _, p := range thread.Participants {
		if p.EntityID == agentID && p.EntityType == "user" {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		part := entity.NewThreadParticipant(thread.ID, "user", agentID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return nil, err
		}
		thread.Participants = append(thread.Participants, part)
	}

	// 2. Determine Parent (Reply to last message)
	msgs, err := s.msgRepo.ListByThread(ctx, thread.ID, 1, 0)
	var parentID *uuid.UUID
	if err == nil && len(msgs) > 0 {
		parentID = &msgs[0].ID
	}

	// 3. Create Message
	msg := entity.NewMessage(tenantID, thread.EnvironmentID, thread.ID, "agent", &agentID, content)
	msg.ParentID = parentID

	return s.msgRepo.Create(ctx, msg)
}

func (s *conversationServiceImpl) AddInternalNote(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, authorID uuid.UUID, content []byte) (*entity.Message, error) {
	// 1. Fetch Thread with environment check
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, fmt.Errorf("thread not found or environment mismatch: %w", err)
	}

	// Add Author as Participant if not already
	isParticipant := false
	for _, p := range thread.Participants {
		if p.EntityID == authorID && p.EntityType == "user" {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		part := entity.NewThreadParticipant(thread.ID, "user", authorID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return nil, err
		}
	}

	// 2. Create Internal Note Message
	msg := entity.NewInternalNote(tenantID, thread.EnvironmentID, thread.ID, authorID, content)

	// 3. Save Message
	return s.msgRepo.Create(ctx, msg)
}

func (s *conversationServiceImpl) AssignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, memberID uuid.UUID) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}

	thread.Status = "assigned"
	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return err
	}

	// Add Member as Participant if not already
	isParticipant := false
	for _, p := range thread.Participants {
		if p.EntityID == memberID && p.EntityType == "user" {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		part := entity.NewThreadParticipant(threadID, "user", memberID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return err
		}
	}

	// Create Assignment Log
	log := entity.NewAssignmentLog(threadID, &memberID)
	return s.logRepo.Create(ctx, log)
}

func (s *conversationServiceImpl) UnassignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}

	if thread.Status == "resolved" {
		return errors.New("cannot unassign resolved thread")
	}

	// Update Status to unassigned (or open?) - adhering to 'unassigned' as per audit
	thread.Status = "unassigned"

	// Logic to remove 'assigned participant' could be complex if multiple.
	// For now, assume 'AssignThread' adds one. We do not remove participant from list (history),
	// but the status change indicates it's back in pool.
	// Optionally, we could find the current assignee and remove them from 'participants' list or keep them?
	// Recommendation: Keep them as participant (they were part of it), but status drives the inbox view.

	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return err
	}

	// Log the unassignment? Maybe create a log with nil memberID?
	// log := entity.NewAssignmentLog(threadID, nil) // optional
	return nil
}

func (s *conversationServiceImpl) BulkAssignThreads(ctx context.Context, tenantID, envID uuid.UUID, threadIDs []uuid.UUID, memberID uuid.UUID) error {
	for _, threadID := range threadIDs {
		// Use existing single assign logic
		if err := s.AssignThread(ctx, tenantID, envID, threadID, memberID); err != nil {
			// Continue or fail? Bulk usually implies best effort or transaction.
			// Let's log error and continue (standard for non-transactional bulk APIs) or return first error.
			// For simplicity/safer data, return error.
			return fmt.Errorf("failed to assign thread %s: %w", threadID, err)
		}
	}
	return nil
}

func (s *conversationServiceImpl) ResolveThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}

	thread.Status = "resolved"
	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return err
	}

	// Resolution Analytics
	log, err := s.logRepo.GetLastByThread(ctx, threadID)
	if err == nil && log != nil {
		now := time.Now()
		log.ResolvedAt = &now
		duration := int(now.Sub(log.AssignedAt).Seconds())
		log.ResponseTimeSeconds = &duration
		return s.logRepo.Update(ctx, log)
	}

	return nil
}

func (s *conversationServiceImpl) GetThread(ctx context.Context, envID, threadID uuid.UUID) (*entity.Thread, error) {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, err
	}

	// Enrich with participants
	parts, err := s.threadRepo.GetParticipants(ctx, thread.ID)
	if err != nil {
		return nil, err
	}
	thread.Participants = parts

	return thread, nil
}

func (s *conversationServiceImpl) GetMessagesByThread(ctx context.Context, envID, threadID uuid.UUID, limit, offset int) ([]*entity.Message, int64, error) {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, 0, err
	}
	msgs, err := s.msgRepo.ListByThread(ctx, threadID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	// Total count could be added if repository supports it, for now returning len and error
	return msgs, int64(len(msgs)), nil
}

func (s *conversationServiceImpl) GetMessagesByCursor(ctx context.Context, envID, threadID uuid.UUID, cursorStr string, direction string, limit int) ([]*entity.Message, string, string, error) {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, "", "", err
	}

	var cQuery *cursor.Cursor
	if cursorStr != "" {
		t, id, err := cursor.Decode(cursorStr)
		if err != nil {
			return nil, "", "", fmt.Errorf("invalid cursor: %w", err)
		}
		cQuery = &cursor.Cursor{Timestamp: t.UnixMicro(), ID: id}
	}

	// Always fetch limit + 1 to check if there are more items
	fetchLimit := limit + 1

	msgs, err := s.msgRepo.ListByCursor(ctx, threadID, cQuery, direction, fetchLimit)
	if err != nil {
		return nil, "", "", err
	}

	// Determine pagination meta
	var nextCursorStr, prevCursorStr string
	hasMore := len(msgs) > limit

	if hasMore {
		// Remove the extra item
		msgs = msgs[:limit]
	}

	if len(msgs) > 0 {
		// Logic to generate next/prev cursors:
		// Based on direction and sort order.
		// Current repo implementation:
		// - 'after' (newer): ASC (oldest -> newest)
		// - 'before' (older, default): DESC (newest -> oldest)

		if direction == "after" {
			// Result: [Older ... Newer]
			// Next (Older than first) -> Prev Page logic
			// Prev (Newer than last) -> Next Page logic relative to direction??
			// Let's stick to standard:
			// next_cursor: to load MORE in the SAME direction
			// prev_cursor: to load back in OPPOSITE direction?

			// Just return the boundary cursors
			first := msgs[0]
			last := msgs[len(msgs)-1]

			// Because sorted ASC: First is oldest in this batch, Last is newest in this batch
			// To get OLDER than First -> direction='before', cursor=First
			// To get NEWER than Last -> direction='after', cursor=Last
			prevCursorStr = cursor.Encode(first.CreatedAt, first.ID)
			if hasMore {
				nextCursorStr = cursor.Encode(last.CreatedAt, last.ID)
			}
		} else {
			// Default 'before' (older)
			// Result: [Newest ... Oldest]
			first := msgs[0]
			last := msgs[len(msgs)-1]

			// First is Newest, Last is Oldest
			// To get NEWER than First -> direction='after', cursor=First
			// To get OLDER than Last -> direction='before', cursor=Last
			prevCursorStr = cursor.Encode(first.CreatedAt, first.ID)
			if hasMore {
				nextCursorStr = cursor.Encode(last.CreatedAt, last.ID)
			}
		}
	}

	return msgs, nextCursorStr, prevCursorStr, nil
}

func (s *conversationServiceImpl) GetOrCreateDirectThread(ctx context.Context, envID, memberID uuid.UUID, targetID uuid.UUID, targetType string) (*entity.Thread, error) {
	var thread *entity.Thread
	var err error

	// 1. Generate ReferenceHash for Direct Chat (Sorted IDs)
	ids := []string{memberID.String(), targetID.String()}
	if memberID.String() > targetID.String() {
		ids[0], ids[1] = ids[1], ids[0]
	}
	refHash := fmt.Sprintf("direct:%s:%s", ids[0], ids[1])

	// 2. Try to create new thread (Atomic with unique constraint)
	thread = entity.NewThread(envID, entity.ThreadTypeDirect, "internal")
	thread.ReferenceHash = &refHash
	thread, err = s.threadRepo.Create(ctx, thread)
	if err != nil {
		// 3. If failed (Unique Violation), try to fetch existing
		thread, err = s.threadRepo.GetDirectThreadBetweenEntities(ctx, "user", memberID, targetType, targetID)
		if err != nil {
			return nil, err
		}
		return thread, nil
	}

	// 3. Add Participants
	p1 := entity.NewThreadParticipant(thread.ID, "user", memberID)
	p2 := entity.NewThreadParticipant(thread.ID, targetType, targetID)

	if err := s.threadRepo.AddParticipant(ctx, p1); err != nil {
		return nil, err
	}
	if err := s.threadRepo.AddParticipant(ctx, p2); err != nil {
		return nil, err
	}

	thread.Participants = []*entity.ThreadParticipant{p1, p2}
	return thread, nil
}

func (s *conversationServiceImpl) CreateGroupThread(ctx context.Context, envID uuid.UUID, name string, inputParticipants []*entity.ThreadParticipant) (*entity.Thread, error) {
	thread := entity.NewThread(envID, entity.ThreadTypeGroup, "internal")
	// Metadata can store name if needed
	meta := map[string]interface{}{
		"name": name,
	}
	metaBytes, _ := json.Marshal(meta)
	thread.Metadata = metaBytes

	thread, err := s.threadRepo.Create(ctx, thread)
	if err != nil {
		return nil, err
	}

	var savedParticipants []*entity.ThreadParticipant
	for _, p := range inputParticipants {
		// Create new participant object with correct ThreadID
		part := entity.NewThreadParticipant(thread.ID, p.EntityType, p.EntityID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return nil, err
		}
		savedParticipants = append(savedParticipants, part)
	}
	thread.Participants = savedParticipants

	return thread, nil
}

func (s *conversationServiceImpl) UpdateGroupThread(ctx context.Context, envID, threadID uuid.UUID, name string) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}

	if thread.Type != entity.ThreadTypeGroup {
		return errors.New("not a group thread")
	}

	meta := map[string]interface{}{}
	if err := json.Unmarshal(thread.Metadata, &meta); err != nil {
		meta = map[string]interface{}{}
	}
	meta["name"] = name
	metaBytes, _ := json.Marshal(meta)
	thread.Metadata = metaBytes

	return s.threadRepo.Update(ctx, thread)
}

func (s *conversationServiceImpl) AddGroupParticipants(ctx context.Context, envID, threadID uuid.UUID, inputParticipants []*entity.ThreadParticipant) error {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}
	for _, p := range inputParticipants {
		part := entity.NewThreadParticipant(threadID, p.EntityType, p.EntityID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return err
		}
	}
	return nil
}

func (s *conversationServiceImpl) RemoveGroupParticipant(ctx context.Context, envID, threadID uuid.UUID, entityType string, entityID uuid.UUID) error {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}
	return s.threadRepo.RemoveParticipant(ctx, threadID, entityType, entityID)
}

func (s *conversationServiceImpl) MarkThreadRead(ctx context.Context, envID, threadID, memberID uuid.UUID) error {
	// 1. Find the participant record with security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return err
	}

	parts, err := s.threadRepo.GetParticipants(ctx, threadID)
	if err != nil {
		return err
	}

	var targetPart *entity.ThreadParticipant
	for _, p := range parts {
		if p.EntityID == memberID && p.EntityType == "user" {
			targetPart = p
			break
		}
	}

	if targetPart == nil {
		return errors.New("participant not found")
	}

	now := time.Now()
	targetPart.LastReadAt = &now
	return s.threadRepo.UpdateParticipant(ctx, targetPart)
}

func (s *conversationServiceImpl) ListThreads(ctx context.Context, tenantID, envID uuid.UUID, status string, assignedToMe bool, memberID uuid.UUID, limit, offset int) ([]*entity.Thread, int64, error) {
	// Prepare Filter
	var statusPtr *string
	if status != "" && status != "all" {
		statusPtr = &status
	}

	var memberPtr *uuid.UUID
	if memberID != uuid.Nil {
		memberPtr = &memberID
	}

	var envPtr *uuid.UUID
	if envID != uuid.Nil {
		envPtr = &envID
	}

	filter := domainRepo.ThreadFilter{
		Status:        statusPtr,
		AssignedToMe:  assignedToMe,
		MemberID:      memberPtr,
		EnvironmentID: envPtr,
		Limit:         limit,
		Offset:        offset,
	}

	threads, total, err := s.threadRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Enrich with participants
	for _, t := range threads {
		parts, _ := s.threadRepo.GetParticipants(ctx, t.ID)
		t.Participants = parts
	}

	return threads, total, nil
}

func (s *conversationServiceImpl) GetTeamDashboard(ctx context.Context, req dto.DashboardStatsRequest) (*dto.TeamDashboardResponse, error) {
	from, to := s.getPeriod(req.From, req.To)
	sla := s.getSLA(req.SLAThreshold)

	stats, err := s.logRepo.GetTeamStats(ctx, req.EnvironmentID, from, to, sla)
	if err != nil {
		return nil, err
	}

	return &dto.TeamDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetPartnerDashboard(ctx context.Context, req dto.DashboardStatsRequest) (*dto.PartnerDashboardResponse, error) {
	// For now, Partner Dashboard structure is similar to Team
	from, to := s.getPeriod(req.From, req.To)
	sla := s.getSLA(req.SLAThreshold)

	stats, err := s.logRepo.GetTeamStats(ctx, req.EnvironmentID, from, to, sla)
	if err != nil {
		return nil, err
	}

	return &dto.PartnerDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetAgentDashboard(ctx context.Context, memberID uuid.UUID, req dto.DashboardStatsRequest) (*dto.AgentDashboardResponse, error) {
	from, to := s.getPeriod(req.From, req.To)
	sla := s.getSLA(req.SLAThreshold)

	stats, err := s.logRepo.GetAgentStats(ctx, req.EnvironmentID, memberID, from, to, sla)
	if err != nil {
		return nil, err
	}

	return &dto.AgentDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetTeamStats(ctx context.Context, tenantID, envID uuid.UUID, from, to time.Time) (*entity.TeamStats, error) {
	return s.logRepo.GetTeamStats(ctx, envID, from, to, 900) // Default 15m
}

func (s *conversationServiceImpl) GetAgentStats(ctx context.Context, tenantID, envID uuid.UUID, memberID uuid.UUID, from, to time.Time) (*entity.AgentStats, error) {
	return s.logRepo.GetAgentStats(ctx, envID, memberID, from, to, 900) // Default 15m
}

func (s *conversationServiceImpl) getPeriod(from, to time.Time) (time.Time, time.Time) {
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.Add(-24 * time.Hour)
	}
	return from, to
}

func (s *conversationServiceImpl) getSLA(threshold int) int {
	if threshold <= 0 {
		return 900 // Default 15m
	}
	return threshold
}
