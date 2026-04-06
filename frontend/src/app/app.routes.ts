import { Routes } from '@angular/router';
import { authGuard } from './guards/auth-guard';

export const routes: Routes = [
    { path: 'login', loadComponent: () => import('./pages/login/login').then((m) => m.LoginPage) },
    {
        path: 'register',
        loadComponent: () => import('./pages/register/register').then((m) => m.RegisterPage),
    },
    {
        path: 'dashboard',
        loadComponent: () => import('./pages/dashboard/dashboard').then((m) => m.DashboardPage),
        canActivate: [authGuard],
    },
    {
        path: 'accounts',
        loadComponent: () => import('./pages/accounts/accounts').then((m) => m.AccountsPage),
        canActivate: [authGuard],
    },
    {
        path: 'budgets',
        loadComponent: () => import('./pages/budgets/budgets').then((m) => m.BudgetsPage),
        canActivate: [authGuard],
    },
    {
        path: 'transactions',
        loadComponent: () =>
            import('./pages/transactions/transactions').then((m) => m.TransactionsPage),
        canActivate: [authGuard],
    },
    {
        path: 'reports',
        loadComponent: () => import('./pages/reports/reports').then((m) => m.ReportsPage),
        canActivate: [authGuard],
    },
    { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
    { path: '**', redirectTo: 'dashboard' },
];
