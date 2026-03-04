import { Component, inject, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth';

@Component({
    selector: 'app-login',
    imports: [FormsModule, RouterLink],
    templateUrl: './login.html',
    styleUrl: './login.css',
})
export class LoginPage {
    private auth = inject(AuthService);
    private router = inject(Router);

    email = '';
    password = '';
    error: WritableSignal<string> = signal('');
    loading: WritableSignal<boolean> = signal(false);

    onSubmit() {
        this.error.set('');
        this.loading.set(true);
        this.auth.login(this.email, this.password).subscribe({
            next: () => {
                this.loading.set(false);
                this.router.navigate(['/dashboard']);
            },
            error: () => {
                this.loading.set(false);
                this.error.set('Invalid email or password.');
            },
        });
    }
}
