import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { BehaviorSubject, catchError, map, Observable, of, tap } from 'rxjs';
import { environment } from '../../environments/environment';
import { AuthResponse } from '../models/models';

const TOKEN_KEY = 'budget_jwt';

@Injectable({ providedIn: 'root' })
export class AuthService {
    private http = inject(HttpClient);
    private router = inject(Router);
    private loggedIn$ = new BehaviorSubject<boolean>(this.hasToken());

    get isLoggedIn$(): Observable<boolean> {
        return this.loggedIn$.asObservable();
    }

    get isLoggedIn(): boolean {
        return this.loggedIn$.value;
    }

    getToken(): string | null {
        return localStorage.getItem(TOKEN_KEY);
    }

    login(email: string, password: string): Observable<AuthResponse> {
        return this.http
            .post<AuthResponse>(`${environment.apiUrl}/auth/login`, { email, password })
            .pipe(tap((res) => this.setSession(res)));
    }

    register(email: string, password: string): Observable<AuthResponse> {
        return this.http
            .post<AuthResponse>(`${environment.apiUrl}/auth/register`, { email, password })
            .pipe(tap((res) => this.setSession(res)));
    }

    logout(): void {
        localStorage.removeItem(TOKEN_KEY);
        this.loggedIn$.next(false);
        this.router.navigate(['/login']);
    }

    private setSession(res: AuthResponse): void {
        localStorage.setItem(TOKEN_KEY, res.token);
        this.loggedIn$.next(true);
    }

    private hasToken(): boolean {
        return !!localStorage.getItem(TOKEN_KEY);
    }

    // Add this to your AuthService
checkAuthStatus(): Observable<boolean> {
    const token = this.getToken();
    if (!token) {
        this.loggedIn$.next(false);
        return of(false);
    }

    // Call your 'profile' or 'me' endpoint
    return this.http.get<any>(`${environment.apiUrl}/auth/validate`).pipe(
        map(() => {
            this.loggedIn$.next(true);
            return true;
        }),
        catchError(() => {
            this.logout(); // Clear local storage and state if server rejects token
            return of(false);
        })
    );
}
}
