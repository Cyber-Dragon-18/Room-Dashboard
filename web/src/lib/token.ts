const KEY = 'room-token'

export function getToken(): string | null {
    return localStorage.getItem(KEY);
}

export function saveToken(token: string): void {
    localStorage.setItem(KEY, token.trim());
}

export function clearToken():void {
    localStorage.removeItem(KEY)
}

