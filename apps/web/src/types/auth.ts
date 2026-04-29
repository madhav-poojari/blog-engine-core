export interface Admin {
  id: string
  username: string
  createdAt: string
}

export interface LoginInput {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  admin: Admin
}
