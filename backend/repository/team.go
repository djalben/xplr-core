package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aalabin/xplr/models"
)

// CreateTeam - Создать команду
func CreateTeam(ownerID int, name string) (*models.Team, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var team models.Team
	err := GlobalDB.QueryRow(
		`INSERT INTO teams (name, owner_id) 
		 VALUES ($1, $2) 
		 RETURNING id, name, owner_id, created_at, updated_at`,
		name, ownerID,
	).Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		log.Printf("DB Error creating team: %v", err)
		return nil, fmt.Errorf("failed to create team")
	}

	// Добавить владельца в команду как owner
	_, err = GlobalDB.Exec(
		`INSERT INTO team_members (team_id, user_id, role, invited_by) 
		 VALUES ($1, $2, 'owner', $2)`,
		team.ID, ownerID,
	)
	if err != nil {
		log.Printf("DB Error adding owner to team: %v", err)
		// Откатываем создание команды
		GlobalDB.Exec("DELETE FROM teams WHERE id = $1", team.ID)
		return nil, fmt.Errorf("failed to add owner to team")
	}

	log.Printf("✅ Team %d created successfully by user %d", team.ID, ownerID)
	return &team, nil
}

// GetUserTeams - Получить команды пользователя
func GetUserTeams(userID int) ([]models.Team, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT t.id, t.name, t.owner_id, t.created_at, t.updated_at
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC
	`
	rows, err := GlobalDB.Query(query, userID)
	if err != nil {
		log.Printf("DB Error fetching teams for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to fetch teams")
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning team: %v", err)
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// GetTeam - Получить команду по ID
func GetTeam(teamID int) (*models.Team, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var team models.Team
	err := GlobalDB.QueryRow(
		"SELECT id, name, owner_id, created_at, updated_at FROM teams WHERE id = $1",
		teamID,
	).Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		log.Printf("DB Error fetching team %d: %v", teamID, err)
		return nil, fmt.Errorf("failed to fetch team")
	}

	return &team, nil
}

// GetTeamMembers - Получить участников команды
func GetTeamMembers(teamID int) ([]models.TeamMember, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.invited_by, tm.joined_at,
		       u.id, u.email, u.balance, COALESCE(u.balance_rub, 0), u.status
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at ASC
	`
	rows, err := GlobalDB.Query(query, teamID)
	if err != nil {
		log.Printf("DB Error fetching team members for team %d: %v", teamID, err)
		return nil, fmt.Errorf("failed to fetch team members")
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		var user models.User
		var invitedBy sql.NullInt64

		err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID, &member.Role,
			&invitedBy, &member.JoinedAt,
			&user.ID, &user.Email, &user.Balance, &user.BalanceRub, &user.Status,
		)
		if err != nil {
			log.Printf("Error scanning team member: %v", err)
			continue
		}

		if invitedBy.Valid {
			invitedByVal := int(invitedBy.Int64)
			member.InvitedBy = &invitedByVal
		}

		member.User = &user
		members = append(members, member)
	}

	return members, nil
}

// InviteTeamMember - Пригласить участника в команду
func InviteTeamMember(teamID int, inviterID int, email string, role string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// 1. Проверить права приглашающего (должен быть owner или admin)
	hasAccess, inviterRole, err := CheckTeamAccess(teamID, inviterID)
	if err != nil || !hasAccess {
		return fmt.Errorf("access denied")
	}
	if inviterRole != "owner" && inviterRole != "admin" {
		return fmt.Errorf("insufficient permissions: only owner or admin can invite members")
	}

	// 2. Найти пользователя по email
	user, err := GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("user with email %s not found", email)
	}

	// 3. Проверить, что пользователь еще не в команде
	var existingMemberID int
	err = GlobalDB.QueryRow(
		"SELECT id FROM team_members WHERE team_id = $1 AND user_id = $2",
		teamID, user.ID,
	).Scan(&existingMemberID)
	if err == nil {
		return fmt.Errorf("user is already a member of this team")
	}

	// 4. Валидация роли
	if role != "admin" && role != "member" {
		return fmt.Errorf("invalid role: must be 'admin' or 'member'")
	}

	// 5. Добавить участника
	_, err = GlobalDB.Exec(
		`INSERT INTO team_members (team_id, user_id, role, invited_by) 
		 VALUES ($1, $2, $3, $4)`,
		teamID, user.ID, role, inviterID,
	)
	if err != nil {
		log.Printf("DB Error inviting team member: %v", err)
		return fmt.Errorf("failed to invite team member")
	}

	log.Printf("✅ User %d invited to team %d as %s by user %d", user.ID, teamID, role, inviterID)
	return nil
}

// RemoveTeamMember - Удалить участника из команды
func RemoveTeamMember(teamID int, userID int, removerID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// 1. Проверить права удаляющего
	hasAccess, removerRole, err := CheckTeamAccess(teamID, removerID)
	if err != nil || !hasAccess {
		return fmt.Errorf("access denied")
	}
	if removerRole != "owner" && removerRole != "admin" {
		return fmt.Errorf("insufficient permissions: only owner or admin can remove members")
	}

	// 2. Нельзя удалить владельца команды
	var memberRole string
	err = GlobalDB.QueryRow(
		"SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2",
		teamID, userID,
	).Scan(&memberRole)
	if err != nil {
		return fmt.Errorf("member not found")
	}
	if memberRole == "owner" {
		return fmt.Errorf("cannot remove team owner")
	}

	// 3. Удалить участника
	_, err = GlobalDB.Exec(
		"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
		teamID, userID,
	)
	if err != nil {
		log.Printf("DB Error removing team member: %v", err)
		return fmt.Errorf("failed to remove team member")
	}

	log.Printf("✅ User %d removed from team %d by user %d", userID, teamID, removerID)
	return nil
}

// UpdateTeamMemberRole - Изменить роль участника
func UpdateTeamMemberRole(teamID int, userID int, newRole string, updaterID int) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// 1. Проверить права (только owner может менять роли)
	hasAccess, updaterRole, err := CheckTeamAccess(teamID, updaterID)
	if err != nil || !hasAccess {
		return fmt.Errorf("access denied")
	}
	if updaterRole != "owner" {
		return fmt.Errorf("insufficient permissions: only owner can change roles")
	}

	// 2. Валидация новой роли
	if newRole != "admin" && newRole != "member" {
		return fmt.Errorf("invalid role: must be 'admin' or 'member'")
	}

	// 3. Нельзя изменить роль владельца
	var currentRole string
	err = GlobalDB.QueryRow(
		"SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2",
		teamID, userID,
	).Scan(&currentRole)
	if err != nil {
		return fmt.Errorf("member not found")
	}
	if currentRole == "owner" {
		return fmt.Errorf("cannot change owner role")
	}

	// 4. Обновить роль
	_, err = GlobalDB.Exec(
		"UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3",
		newRole, teamID, userID,
	)
	if err != nil {
		log.Printf("DB Error updating team member role: %v", err)
		return fmt.Errorf("failed to update team member role")
	}

	log.Printf("✅ User %d role updated to %s in team %d by user %d", userID, newRole, teamID, updaterID)
	return nil
}

// CheckTeamAccess - Проверить доступ пользователя к команде
func CheckTeamAccess(teamID int, userID int) (bool, string, error) {
	if GlobalDB == nil {
		return false, "", fmt.Errorf("database connection not initialized")
	}

	var role string
	err := GlobalDB.QueryRow(
		"SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2",
		teamID, userID,
	).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil // Нет доступа, но это не ошибка
		}
		return false, "", err
	}

	return true, role, nil
}
