// ===== Models =====

export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface Client {
  id: string;
  user_id: string;
  name: string;
  industry: string;
  created_at: string;
  updated_at: string;
}

export interface Outlet {
  id: string;
  user_id: string;
  name: string;
  website: string;
  country: string;
  created_at: string;
  updated_at: string;
}

export interface Journalist {
  id: string;
  user_id: string;
  outlet_id: string;
  outlet: Outlet;
  name: string;
  email: string;
  niche: string;
  created_at: string;
  updated_at: string;
}

export interface CrmRelationship {
  id: string;
  user_id: string;
  journalist_id: string;
  journalist: Journalist;
  relationship_score: number;
  private_notes: string;
  created_at: string;
  updated_at: string;
}

export interface Campaign {
  id: string;
  user_id: string;
  client_id: string;
  client: Client;
  title: string;
  press_release_text: string;
  created_at: string;
  updated_at: string;
}

export interface Pitch {
  id: string;
  user_id: string;
  client_id: string;
  client: Client;
  campaign_id: string | null;
  campaign: Campaign | null;
  journalist_id: string;
  journalist: Journalist;
  context_brief: string;
  status: PitchStatus;
  created_at: string;
  updated_at: string;
}

export interface PitchVersion {
  id: string;
  pitch_id: string;
  version_number: number;
  ai_generated_body: string;
  prompt_snapshot: string;
  created_at: string;
}

export type PitchStatus = "draft" | "sent" | "opened" | "replied";

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

// ===== Request types =====

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface CreateClientRequest {
  name: string;
  industry: string;
}

export interface UpdateClientRequest {
  name?: string;
  industry?: string;
}

export interface CreateOutletRequest {
  name: string;
  website?: string;
  country?: string;
}

export interface UpdateOutletRequest {
  name?: string;
  website?: string;
  country?: string;
}

export interface CreateJournalistRequest {
  outlet_id: string;
  name: string;
  email: string;
  niche?: string;
}

export interface UpdateJournalistRequest {
  outlet_id?: string;
  name?: string;
  email?: string;
  niche?: string;
}

export interface CreateCrmRelationshipRequest {
  journalist_id: string;
  relationship_score: number;
  private_notes?: string;
}

export interface UpdateCrmRelationshipRequest {
  relationship_score?: number;
  private_notes?: string;
}

export interface CreateCampaignRequest {
  client_id: string;
  title: string;
  press_release_text?: string;
}

export interface UpdateCampaignRequest {
  title?: string;
  press_release_text?: string;
}

export interface CreatePitchRequest {
  client_id: string;
  journalist_id: string;
  campaign_id?: string;
  context_brief?: string;
}

export interface UpdatePitchRequest {
  status?: PitchStatus;
  context_brief?: string;
}

// ===== Response types =====

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ErrorResponse {
  error: string;
}
