package impl

import (
	"CONVERDA/internal/messaging/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/messaging/application/service"
	"CONVERDA/internal/messaging/controller/dto"
	"CONVERDA/internal/messaging/domain/model/entity"
	"CONVERDA/internal/messaging/domain/repository"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/internal/messaging/infrastructure/gateway"
	"CONVERDA/pkg/cursor"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type conversationServiceImpl struct {
	msgRepo      domainRepo.MessageRepository
	threadRepo   domainRepo.ThreadRepository
	subRepo      domainRepo.SubscriberRepository
	logRepo      domainRepo.AssignmentLogRepository
	memberReader domainRepo.MemberReader
	appReader    domainRepo.AppReader
	hub          *gateway.Hub
}

func NewConversationService(
	msgRepo domainRepo.MessageRepository,
	threadRepo domainRepo.ThreadRepository,
	subRepo domainRepo.SubscriberRepository,
	logRepo domainRepo.AssignmentLogRepository,
	memberReader domainRepo.MemberReader,
	appReader domainRepo.AppReader,
	hub *gateway.Hub,
) service.ConversationService {
	return &conversationServiceImpl{
		msgRepo:      msgRepo,
		threadRepo:   threadRepo,
		subRepo:      subRepo,
		logRepo:      logRepo,
		memberReader: memberReader,
		appReader:    appReader,
		hub:          hub,
	}
}

func (s *conversationServiceImpl) ReceiveMessage(ctx context.Context, tenantID, envID uuid.UUID, subKey, channel string, content []byte, senderID *uuid.UUID) (*entity.Message, error) {
	// 1. Find or Create Subscriber
	sub, err := s.subRepo.GetByKey(ctx, envID, subKey)
	if err != nil {
		sub = entity.NewSubscriber(envID, subKey)
		sub, err = s.subRepo.Create(ctx, sub)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscriber: %w", err)
		}
	}

	// 2. Find Latest Support Thread for this subscriber
	supportType := "support"
	threads, _, err := s.threadRepo.List(ctx, domainRepo.ThreadFilter{
		EnvironmentID: &envID,
		Type:          &supportType,
		MemberID:      &sub.ID,
		Limit:         1,
	})

	var thread *entity.Thread
	if len(threads) > 0 {
		thread = threads[0]
	}

	if thread == nil {
		// Create new support thread
		if channel == "" {
			channel = "support"
		}
		thread = entity.NewThread(envID, "support", channel)
		thread, err = s.threadRepo.Create(ctx, thread)
		if err != nil {
			return nil, fmt.Errorf("failed to create support thread: %w", err)
		}
		// Add Subscriber as Participant
		part := entity.NewThreadParticipant(thread.ID, "subscriber", sub.ID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return nil, fmt.Errorf("failed to add subscriber participant: %w", err)
		}
		thread.Participants = append(thread.Participants, part)
	} else if thread.Status == entity.ThreadStatusResolved {
		// Reopen resolved thread
		if err := thread.TransitionTo(entity.ThreadStatusUnassigned); err != nil {
			return nil, err
		}
		thread.IsOverdue = false
		if err := s.threadRepo.Update(ctx, thread); err != nil {
			return nil, fmt.Errorf("failed to reopen thread: %w", err)
		}
		s.broadcastEvent(envID, "thread_reopened", map[string]interface{}{
			"threadID": thread.ID,
		})
	}

	// 3. Add Message
	realSenderID := sub.ID
	if senderID != nil {
		realSenderID = *senderID
	}

	msg := entity.NewMessage(tenantID, envID, thread.ID, "contact", &realSenderID, content)
	createdMsg, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	s.broadcastEvent(envID, "message_received", createdMsg)
	return createdMsg, nil
}

func (s *conversationServiceImpl) ReplyMessage(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, agentID uuid.UUID, content []byte) (*entity.Message, error) {
	// 1. Fetch Thread with environment check
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, fmt.Errorf("thread not found or environment mismatch: %w", err)
	}

	if thread.Status == entity.ThreadStatusResolved {
		return nil, domain.ErrThreadResolved
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
			return nil, fmt.Errorf("failed to add agent as participant: %w", err)
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

	createdMsg, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply message: %w", err)
	}
	s.broadcastEvent(envID, "message_replied", createdMsg)
	return createdMsg, nil
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
			return nil, fmt.Errorf("failed to add author as participant: %w", err)
		}
	}

	// 2. Create Internal Note Message
	msg := entity.NewInternalNote(tenantID, thread.EnvironmentID, thread.ID, authorID, content)

	// 3. Save Message
	createdNote, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal note: %w", err)
	}
	return createdNote, nil
}

func (s *conversationServiceImpl) AssignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, memberID uuid.UUID) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread %s: %w", threadID, err)
	}

	if err := thread.TransitionTo(entity.ThreadStatusAssigned); err != nil {
		return fmt.Errorf("failed to transition thread to assigned: %w", err)
	}
	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return fmt.Errorf("failed to update thread %s: %w", threadID, err)
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
			return fmt.Errorf("failed to add member as participant: %w", err)
		}
	}

	// Create Assignment Log
	log := entity.NewAssignmentLog(threadID, &memberID)
	if err := s.logRepo.Create(ctx, log); err != nil {
		return fmt.Errorf("failed to create assignment log: %w", err)
	}

	s.broadcastEvent(envID, "thread_assigned", map[string]interface{}{
		"threadID":   threadID,
		"assignedTo": memberID,
	})

	return nil
}

func (s *conversationServiceImpl) UnassignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID) error {
	thread, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread %s: %w", threadID, err)
	}

	if thread.Status == entity.ThreadStatusResolved {
		return domain.ErrThreadResolved
	}

	if err := thread.TransitionTo(entity.ThreadStatusUnassigned); err != nil {
		return fmt.Errorf("failed to transition thread to unassigned: %w", err)
	}

	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return fmt.Errorf("failed to update thread %s: %w", threadID, err)
	}

	s.broadcastEvent(envID, "thread_unassigned", map[string]interface{}{
		"threadID": threadID,
	})
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
		return fmt.Errorf("failed to fetch thread %s: %w", threadID, err)
	}

	if err := thread.TransitionTo(entity.ThreadStatusResolved); err != nil {
		return fmt.Errorf("failed to transition thread to resolved: %w", err)
	}
	thread.IsOverdue = false
	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return fmt.Errorf("failed to update thread %s: %w", threadID, err)
	}

	s.broadcastEvent(envID, "thread_resolved", map[string]interface{}{
		"threadID": threadID,
	})

	// Resolution Analytics
	log, err := s.logRepo.GetLastByThread(ctx, threadID)
	if err == nil && log != nil {
		now := time.Now()
		log.ResolvedAt = &now
		duration := int(now.Sub(log.AssignedAt).Seconds())
		log.ResponseTimeSeconds = &duration
		if err := s.logRepo.Update(ctx, log); err != nil {
			global.Logger.Warn("messaging_service: failed to update resolution log", zap.String("threadID", threadID.String()), zap.Error(err))
		}
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
		return nil, fmt.Errorf("failed to get participants for thread %s: %w", thread.ID, err)
	}
	thread.Participants = parts

	return thread, nil
}

func (s *conversationServiceImpl) GetMessagesByThread(ctx context.Context, envID, threadID uuid.UUID, limit, offset int) ([]*entity.Message, int64, error) {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, 0, fmt.Errorf("thread not found or access denied: %w", err)
	}
	msgs, err := s.msgRepo.ListByThread(ctx, threadID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages for thread %s: %w", threadID, err)
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
		return nil, "", "", fmt.Errorf("failed to list messages by cursor for thread %s: %w", threadID, err)
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
			return nil, fmt.Errorf("failed to get existing direct thread: %w", err)
		}
		return thread, nil
	}

	// 3. Add Participants
	p1 := entity.NewThreadParticipant(thread.ID, "user", memberID)
	p2 := entity.NewThreadParticipant(thread.ID, targetType, targetID)

	if err := s.threadRepo.AddParticipant(ctx, p1); err != nil {
		return nil, fmt.Errorf("failed to add participant 1: %w", err)
	}
	if err := s.threadRepo.AddParticipant(ctx, p2); err != nil {
		return nil, fmt.Errorf("failed to add participant 2: %w", err)
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
		return nil, fmt.Errorf("failed to create group thread: %w", err)
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
		return fmt.Errorf("failed to fetch thread %s: %w", threadID, err)
	}

	if thread.Type != entity.ThreadTypeGroup {
		return domain.ErrNotGroupThread
	}

	meta := map[string]interface{}{}
	if err := json.Unmarshal(thread.Metadata, &meta); err != nil {
		meta = map[string]interface{}{}
	}
	meta["name"] = name
	metaBytes, _ := json.Marshal(meta)
	thread.Metadata = metaBytes

	if err := s.threadRepo.Update(ctx, thread); err != nil {
		return fmt.Errorf("failed to update group thread %s: %w", threadID, err)
	}

	s.broadcastEvent(envID, "thread_updated", map[string]interface{}{
		"threadID": threadID,
		"name":     name,
	})

	return nil
}

func (s *conversationServiceImpl) AddGroupParticipants(ctx context.Context, envID, threadID uuid.UUID, inputParticipants []*entity.ThreadParticipant) error {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread for add participants: %w", err)
	}
	for _, p := range inputParticipants {
		part := entity.NewThreadParticipant(threadID, p.EntityType, p.EntityID)
		if err := s.threadRepo.AddParticipant(ctx, part); err != nil {
			return fmt.Errorf("failed to add participant %s to thread %s: %w", p.EntityID, threadID, err)
		}
	}
	return nil
}

func (s *conversationServiceImpl) RemoveGroupParticipant(ctx context.Context, envID, threadID uuid.UUID, entityType string, entityID uuid.UUID) error {
	// Security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return fmt.Errorf("thread not found or access denied: %w", err)
	}
	if err := s.threadRepo.RemoveParticipant(ctx, threadID, entityType, entityID); err != nil {
		return fmt.Errorf("failed to remove participant %s from thread %s: %w", entityID, threadID, err)
	}
	return nil
}

func (s *conversationServiceImpl) MarkThreadRead(ctx context.Context, envID, threadID, memberID uuid.UUID) error {
	// 1. Find the participant record with security check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return fmt.Errorf("failed to fetch thread for mark read: %w", err)
	}

	parts, err := s.threadRepo.GetParticipants(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get participants for thread %s: %w", threadID, err)
	}

	var targetPart *entity.ThreadParticipant
	for _, p := range parts {
		if p.EntityID == memberID && p.EntityType == "user" {
			targetPart = p
			break
		}
	}

	if targetPart == nil {
		return domain.ErrParticipantNotFound
	}

	now := time.Now()
	targetPart.LastReadAt = &now
	if err := s.threadRepo.UpdateParticipant(ctx, targetPart); err != nil {
		return fmt.Errorf("failed to update participant read status for thread %s: %w", threadID, err)
	}
	return nil
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
		return nil, 0, fmt.Errorf("failed to list threads: %w", err)
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
	sla := req.SLAThreshold
	if sla <= 0 {
		sla = s.getSLAThreshold(ctx, req.EnvironmentID)
	}

	stats, err := s.logRepo.GetTeamStats(ctx, req.EnvironmentID, from, to, sla)
	if err != nil {
		return nil, err
	}

	return &dto.TeamDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetPartnerDashboard(ctx context.Context, req dto.DashboardStatsRequest) (*dto.PartnerDashboardResponse, error) {
	// For now, Partner Dashboard structure is similar to Team
	from, to := s.getPeriod(req.From, req.To)
	sla := req.SLAThreshold
	if sla <= 0 {
		sla = s.getSLAThreshold(ctx, req.EnvironmentID)
	}

	stats, err := s.logRepo.GetTeamStats(ctx, req.EnvironmentID, from, to, sla)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch partner stats: %w", err)
	}

	return &dto.PartnerDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetPersonalDashboard(ctx context.Context, memberID uuid.UUID, req dto.DashboardStatsRequest) (*dto.PersonalDashboardResponse, error) {
	from, to := s.getPeriod(req.From, req.To)

	// 1. Get Agent Stats
	agentStats, err := s.GetAgentStats(ctx, uuid.Nil, req.EnvironmentID, memberID, from, to)
	if err != nil {
		return nil, err
	}

	// 2. Get Team Stats (Comparison)
	teamStats, err := s.GetTeamStats(ctx, uuid.Nil, req.EnvironmentID, from, to)
	if err != nil {
		return nil, err
	}

	// 3. Activity Timeline
	timeline := s.calculateTimeline(ctx, memberID, req.EnvironmentID, from, to)

	return &dto.PersonalDashboardResponse{
		Stats:            agentStats,
		ActivityTimeline: timeline,
		TeamComparison:   teamStats,
	}, nil
}

func (s *conversationServiceImpl) calculateTimeline(ctx context.Context, memberID, envID uuid.UUID, from, to time.Time) []dto.ActivityPoint {
	// Auto-select bucket interval: hourly for ≤24h, daily for longer ranges
	duration := to.Sub(from)
	interval := "hour"
	if duration > 24*time.Hour {
		interval = "day"
	}

	entityPoints, err := s.logRepo.GetActivityTimeline(ctx, envID, memberID, from, to, interval)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("PersonalDashboard: failed to fetch activity timeline: %v", err))
		return []dto.ActivityPoint{}
	}

	// Map entity -> DTO
	points := make([]dto.ActivityPoint, len(entityPoints))
	for i, ep := range entityPoints {
		points[i] = dto.ActivityPoint{
			Time:  ep.Timestamp,
			Value: ep.Value,
		}
	}

	return points
}

func (s *conversationServiceImpl) GetAgentDashboard(ctx context.Context, memberID uuid.UUID, req dto.DashboardStatsRequest) (*dto.AgentDashboardResponse, error) {
	from, to := s.getPeriod(req.From, req.To)
	sla := req.SLAThreshold
	if sla <= 0 {
		sla = s.getSLAThreshold(ctx, req.EnvironmentID)
	}

	stats, err := s.logRepo.GetAgentStats(ctx, req.EnvironmentID, memberID, from, to, sla)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent stats: %w", err)
	}

	return &dto.AgentDashboardResponse{Stats: stats}, nil
}

func (s *conversationServiceImpl) GetTeamStats(ctx context.Context, tenantID, envID uuid.UUID, from, to time.Time) (*entity.TeamStats, error) {
	sla := s.getSLAThreshold(ctx, envID)
	return s.logRepo.GetTeamStats(ctx, envID, from, to, sla)
}

func (s *conversationServiceImpl) GetAgentStats(ctx context.Context, tenantID, envID uuid.UUID, memberID uuid.UUID, from, to time.Time) (*entity.AgentStats, error) {
	sla := s.getSLAThreshold(ctx, envID)
	return s.logRepo.GetAgentStats(ctx, envID, memberID, from, to, sla)
}

func (s *conversationServiceImpl) getSLAThreshold(ctx context.Context, envID uuid.UUID) int {
	if envID == uuid.Nil {
		return 900
	}

	env, err := s.appReader.GetEnvironment(ctx, envID)
	if err == nil && env != nil {
		if env.SLAThresholdSeconds > 0 {
			return env.SLAThresholdSeconds
		}
		// Fallback to app SLA
		app, err := s.appReader.GetApp(ctx, env.AppID)
		if err == nil && app != nil && app.SLAThresholdSeconds > 0 {
			return app.SLAThresholdSeconds
		}
	}

	return 900 // Default 15m
}

func (s *conversationServiceImpl) GetThreadAuditTrail(ctx context.Context, envID, threadID uuid.UUID, page, pageSize int, from, to *time.Time) (*dto.AuditTrailResponse, error) {
	// 1. Security Check
	_, err := s.threadRepo.GetByIDAndEnv(ctx, threadID, envID)
	if err != nil {
		return nil, fmt.Errorf("thread not found or access denied: %w", err)
	}

	// 2. Fetch Logs (paginated)
	filter := repository.AuditTrailFilter{
		ThreadID: threadID,
		From:     from,
		To:       to,
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
	}
	logs, total, err := s.logRepo.ListByThread(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 3. Enrich Names
	nameMap := make(map[uuid.UUID]string)

	enrichedLogs := make([]*dto.AssignmentLogResponse, len(logs))
	for i, l := range logs {
		enrichedLogs[i] = &dto.AssignmentLogResponse{
			ID:                  l.ID,
			ThreadID:            l.ThreadID,
			AssignedToMemberID:  l.AssignedToMemberID,
			AssignedAt:          l.AssignedAt,
			ResolvedAt:          l.ResolvedAt,
			ResponseTimeSeconds: l.ResponseTimeSeconds,
		}

		if l.AssignedToMemberID != nil {
			memberID := *l.AssignedToMemberID
			if name, ok := nameMap[memberID]; ok {
				enrichedLogs[i].AssignedToDisplayName = name
			} else {
				member, err := s.memberReader.GetMember(ctx, memberID)
				if err == nil && member != nil && member.DisplayName != "" {
					nameMap[memberID] = member.DisplayName
					enrichedLogs[i].AssignedToDisplayName = member.DisplayName
				}
			}
		}
	}

	return &dto.AuditTrailResponse{
		ThreadID: threadID,
		Logs:     enrichedLogs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
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
func (s *conversationServiceImpl) broadcastEvent(envID uuid.UUID, eventType string, payload interface{}) {
	if s.hub == nil {
		return
	}

	event := dto.WsEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		global.Logger.Error("messaging_service: failed to marshal broadcast event", zap.String("type", eventType), zap.Error(err))
		return
	}

	s.hub.BroadcastToEnvironment(envID, data)
}
