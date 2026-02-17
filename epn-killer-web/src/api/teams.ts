import apiClient from './axios';

export interface Team {
  id: number;
  name: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  id: number;
  team_id: number;
  user_id: number;
  role: 'owner' | 'admin' | 'member';
  invited_by?: number;
  joined_at: string;
  user?: {
    id: number;
    email: string;
    balance: string;
    status: string;
  };
}

export interface TeamDetail {
  team: Team;
  members: TeamMember[];
}

export interface CreateTeamRequest {
  name: string;
}

export interface InviteTeamMemberRequest {
  email: string;
  role: 'admin' | 'member';
}

// Создать команду
export const createTeam = async (name: string): Promise<Team> => {
  const response = await apiClient.post<Team>('/user/teams', { name });
  return response.data;
};

// Получить команды пользователя
export const getUserTeams = async (): Promise<Team[]> => {
  const response = await apiClient.get<Team[]>('/user/teams');
  return response.data;
};

// Получить детали команды
export const getTeam = async (teamId: number): Promise<TeamDetail> => {
  const response = await apiClient.get<TeamDetail>(`/user/teams/${teamId}`);
  return response.data;
};

// Пригласить участника
export const inviteTeamMember = async (teamId: number, email: string, role: string): Promise<void> => {
  await apiClient.post(`/user/teams/${teamId}/members`, { email, role });
};

// Удалить участника
export const removeTeamMember = async (teamId: number, userId: number): Promise<void> => {
  await apiClient.delete(`/user/teams/${teamId}/members/${userId}`);
};

// Изменить роль участника
export const updateTeamMemberRole = async (teamId: number, userId: number, role: string): Promise<void> => {
  await apiClient.patch(`/user/teams/${teamId}/members/${userId}/role`, { role });
};
