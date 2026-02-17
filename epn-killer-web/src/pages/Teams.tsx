import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../api/axios';
import { getUserTeams, createTeam, getTeam, inviteTeamMember, removeTeamMember, updateTeamMemberRole, Team, TeamDetail } from '../api/teams';
import SidebarLayout from '../components/SidebarLayout';

const Teams: React.FC = () => {
  const navigate = useNavigate();
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<TeamDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [newTeamName, setNewTeamName] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member'>('member');
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  useEffect(() => {
    fetchTeams();
  }, []);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 3500);
    return () => clearTimeout(t);
  }, [toast]);

  const fetchTeams = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        navigate('/login');
        return;
      }
      const data = await getUserTeams();
      setTeams(Array.isArray(data) ? data : []);
      setIsLoading(false);
    } catch (error) {
      console.error('Error fetching teams:', error);
      setIsLoading(false);
    }
  };

  const handleCreateTeam = async () => {
    if (!newTeamName.trim()) {
      setToast({ message: 'Введите название команды', type: 'error' });
      return;
    }

    try {
      const team = await createTeam(newTeamName);
      setToast({ message: 'Команда создана успешно!', type: 'success' });
      setShowCreateModal(false);
      setNewTeamName('');
      await fetchTeams();
    } catch (error: any) {
      console.error('Error creating team:', error);
      setToast({ message: error.response?.data?.message || 'Не удалось создать команду', type: 'error' });
    }
  };

  const handleSelectTeam = async (teamId: number) => {
    try {
      const teamDetail = await getTeam(teamId);
      setSelectedTeam(teamDetail);
    } catch (error) {
      console.error('Error fetching team details:', error);
      setToast({ message: 'Не удалось загрузить детали команды', type: 'error' });
    }
  };

  const handleInviteMember = async () => {
    if (!selectedTeam || !inviteEmail.trim()) {
      setToast({ message: 'Введите email участника', type: 'error' });
      return;
    }

    try {
      await inviteTeamMember(selectedTeam.team.id, inviteEmail, inviteRole);
      setToast({ message: 'Приглашение отправлено!', type: 'success' });
      setShowInviteModal(false);
      setInviteEmail('');
      await handleSelectTeam(selectedTeam.team.id);
    } catch (error: any) {
      console.error('Error inviting member:', error);
      setToast({ message: error.response?.data?.message || 'Не удалось пригласить участника', type: 'error' });
    }
  };

  const handleRemoveMember = async (userId: number) => {
    if (!selectedTeam) return;
    if (!confirm('Вы уверены, что хотите удалить этого участника?')) return;

    try {
      await removeTeamMember(selectedTeam.team.id, userId);
      setToast({ message: 'Участник удален', type: 'success' });
      await handleSelectTeam(selectedTeam.team.id);
    } catch (error: any) {
      console.error('Error removing member:', error);
      setToast({ message: error.response?.data?.message || 'Не удалось удалить участника', type: 'error' });
    }
  };

  const handleUpdateRole = async (userId: number, newRole: string) => {
    if (!selectedTeam) return;

    try {
      await updateTeamMemberRole(selectedTeam.team.id, userId, newRole);
      setToast({ message: 'Роль обновлена', type: 'success' });
      await handleSelectTeam(selectedTeam.team.id);
    } catch (error: any) {
      console.error('Error updating role:', error);
      setToast({ message: error.response?.data?.message || 'Не удалось обновить роль', type: 'error' });
    }
  };

  if (isLoading) {
    return (
      <SidebarLayout>
        <div style={{ padding: '40px', textAlign: 'center', color: '#888c95' }}>Загрузка...</div>
      </SidebarLayout>
    );
  }

  return (
    <SidebarLayout>
    <div style={{ color: '#ffffff' }}>
      {/* Header */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '30px'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <button
            onClick={() => navigate('/dashboard')}
            style={{
              padding: '8px 16px', backgroundColor: 'rgba(255,255,255,0.05)',
              border: '1px solid rgba(255,255,255,0.1)', borderRadius: '8px',
              color: '#888c95', fontSize: '13px', cursor: 'pointer', fontWeight: '600'
            }}>← Назад</button>
          <h1 style={{ margin: 0, fontSize: '24px', fontWeight: '700' }}>Команды</h1>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          style={{
            backgroundColor: '#00e096',
            color: '#000',
            border: 'none',
            padding: '12px 24px',
            borderRadius: '8px',
            fontWeight: '600',
            cursor: 'pointer',
            fontSize: '14px'
          }}>
          + Создать команду
        </button>
      </div>

      {/* Teams List */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: selectedTeam ? '1fr 2fr' : '1fr',
        gap: '20px'
      }}>
        {/* Teams Sidebar */}
        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.03)',
          borderRadius: '12px',
          padding: '20px',
          border: '1px solid rgba(255, 255, 255, 0.1)'
        }}>
          <h3 style={{ margin: '0 0 20px 0', fontSize: '16px', fontWeight: '600' }}>Мои команды</h3>
          {(teams ?? []).length === 0 ? (
            <div style={{ color: '#888c95', textAlign: 'center', padding: '40px 0' }}>
              У вас пока нет команд
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
              {(teams ?? []).map(team => (
                <div
                  key={team.id}
                  onClick={() => handleSelectTeam(team.id)}
                  style={{
                    padding: '15px',
                    backgroundColor: selectedTeam?.team.id === team.id ? 'rgba(0, 224, 150, 0.1)' : 'transparent',
                    border: `1px solid ${selectedTeam?.team.id === team.id ? '#00e096' : 'rgba(255, 255, 255, 0.1)'}`,
                    borderRadius: '8px',
                    cursor: 'pointer',
                    transition: '0.2s'
                  }}
                  onMouseEnter={(e) => {
                    if (selectedTeam?.team.id !== team.id) {
                      e.currentTarget.style.backgroundColor = 'rgba(255, 255, 255, 0.05)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (selectedTeam?.team.id !== team.id) {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }
                  }}>
                  <div style={{ fontWeight: '600', marginBottom: '5px' }}>{team.name}</div>
                  <div style={{ fontSize: '12px', color: '#888c95' }}>
                    ID: {team.id}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Team Details */}
        {selectedTeam && (
          <div style={{
            backgroundColor: 'rgba(255, 255, 255, 0.03)',
            borderRadius: '12px',
            padding: '30px',
            border: '1px solid rgba(255, 255, 255, 0.1)'
          }}>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: '30px'
            }}>
              <div>
                <h2 style={{ margin: '0 0 5px 0', fontSize: '20px', fontWeight: '700' }}>
                  {selectedTeam.team.name}
                </h2>
                <div style={{ fontSize: '12px', color: '#888c95' }}>
                  Создана: {new Date(selectedTeam.team.created_at).toLocaleDateString('ru-RU')}
                </div>
              </div>
              <button
                onClick={() => setShowInviteModal(true)}
                style={{
                  backgroundColor: '#00e096',
                  color: '#000',
                  border: 'none',
                  padding: '10px 20px',
                  borderRadius: '8px',
                  fontWeight: '600',
                  cursor: 'pointer',
                  fontSize: '14px'
                }}>
                + Пригласить
              </button>
            </div>

            <h3 style={{ margin: '0 0 20px 0', fontSize: '16px', fontWeight: '600' }}>Участники</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {(selectedTeam.members ?? []).map(member => (
                <div
                  key={member.id}
                  style={{
                    padding: '15px',
                    backgroundColor: 'rgba(255, 255, 255, 0.05)',
                    borderRadius: '8px',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                  }}>
                  <div>
                    <div style={{ fontWeight: '600', marginBottom: '5px' }}>
                      {member.user?.email || `User ${member.user_id}`}
                    </div>
                    <div style={{ fontSize: '12px', color: '#888c95' }}>
                      Роль: {member.role === 'owner' ? 'Владелец' : member.role === 'admin' ? 'Администратор' : 'Участник'}
                    </div>
                  </div>
                  {member.role !== 'owner' && (
                    <div style={{ display: 'flex', gap: '10px' }}>
                      <select
                        value={member.role}
                        onChange={(e) => handleUpdateRole(member.user_id, e.target.value)}
                        style={{
                          padding: '5px 10px',
                          backgroundColor: 'rgba(255, 255, 255, 0.05)',
                          border: '1px solid rgba(255, 255, 255, 0.1)',
                          borderRadius: '6px',
                          color: '#fff',
                          fontSize: '12px',
                          cursor: 'pointer'
                        }}>
                        <option value="member">Участник</option>
                        <option value="admin">Администратор</option>
                      </select>
                      <button
                        onClick={() => handleRemoveMember(member.user_id)}
                        style={{
                          padding: '5px 10px',
                          backgroundColor: 'rgba(255, 59, 59, 0.2)',
                          border: '1px solid rgba(255, 59, 59, 0.3)',
                          borderRadius: '6px',
                          color: '#ff3b3b',
                          fontSize: '12px',
                          cursor: 'pointer'
                        }}>
                        Удалить
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Create Team Modal */}
      {showCreateModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 10000
        }}
        onClick={() => setShowCreateModal(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)',
            borderRadius: '24px',
            padding: '40px',
            width: '90%',
            maxWidth: '400px',
            border: '1px solid rgba(255, 255, 255, 0.1)'
          }}
          onClick={(e) => e.stopPropagation()}>
            <h2 style={{ margin: '0 0 24px 0', fontSize: '20px', fontWeight: '700' }}>Создать команду</h2>
            <input
              type="text"
              value={newTeamName}
              onChange={(e) => setNewTeamName(e.target.value)}
              placeholder="Название команды"
              style={{
                width: '100%',
                padding: '12px 16px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '12px',
                color: '#fff',
                fontSize: '16px',
                marginBottom: '20px',
                outline: 'none'
              }}
            />
            <div style={{ display: 'flex', gap: '12px' }}>
              <button
                onClick={() => setShowCreateModal(false)}
                style={{
                  flex: 1,
                  padding: '12px',
                  backgroundColor: 'transparent',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '8px',
                  color: '#888c95',
                  cursor: 'pointer'
                }}>
                Отмена
              </button>
              <button
                onClick={handleCreateTeam}
                style={{
                  flex: 1,
                  padding: '12px',
                  backgroundColor: '#00e096',
                  border: 'none',
                  borderRadius: '8px',
                  color: '#000',
                  fontWeight: '600',
                  cursor: 'pointer'
                }}>
                Создать
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Invite Member Modal */}
      {showInviteModal && selectedTeam && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 10000
        }}
        onClick={() => setShowInviteModal(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)',
            borderRadius: '24px',
            padding: '40px',
            width: '90%',
            maxWidth: '400px',
            border: '1px solid rgba(255, 255, 255, 0.1)'
          }}
          onClick={(e) => e.stopPropagation()}>
            <h2 style={{ margin: '0 0 24px 0', fontSize: '20px', fontWeight: '700' }}>Пригласить участника</h2>
            <input
              type="email"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              placeholder="email@example.com"
              style={{
                width: '100%',
                padding: '12px 16px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '12px',
                color: '#fff',
                fontSize: '16px',
                marginBottom: '20px',
                outline: 'none'
              }}
            />
            <select
              value={inviteRole}
              onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member')}
              style={{
                width: '100%',
                padding: '12px 16px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '12px',
                color: '#fff',
                fontSize: '16px',
                marginBottom: '20px',
                outline: 'none',
                cursor: 'pointer'
              }}>
              <option value="member">Участник</option>
              <option value="admin">Администратор</option>
            </select>
            <div style={{ display: 'flex', gap: '12px' }}>
              <button
                onClick={() => setShowInviteModal(false)}
                style={{
                  flex: 1,
                  padding: '12px',
                  backgroundColor: 'transparent',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '8px',
                  color: '#888c95',
                  cursor: 'pointer'
                }}>
                Отмена
              </button>
              <button
                onClick={handleInviteMember}
                style={{
                  flex: 1,
                  padding: '12px',
                  backgroundColor: '#00e096',
                  border: 'none',
                  borderRadius: '8px',
                  color: '#000',
                  fontWeight: '600',
                  cursor: 'pointer'
                }}>
                Пригласить
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Toast */}
      {toast && (
        <div style={{
          position: 'fixed',
          bottom: '20px',
          right: '20px',
          padding: '16px 24px',
          backgroundColor: toast.type === 'success' ? 'rgba(0, 224, 150, 0.2)' : 'rgba(255, 59, 59, 0.2)',
          border: `1px solid ${toast.type === 'success' ? '#00e096' : '#ff3b3b'}`,
          borderRadius: '12px',
          color: toast.type === 'success' ? '#00e096' : '#ff3b3b',
          fontSize: '14px',
          zIndex: 10001
        }}>
          {toast.message}
        </div>
      )}
    </div>
    </SidebarLayout>
  );
};

export default Teams;
