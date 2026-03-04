import { Component, inject, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth';

@Component({
    selector: 'app-register',
    imports: [FormsModule, RouterLink],
    templateUrl: './register.html',
    styleUrl: './register.css',
})
export class RegisterPage {
    private auth = inject(AuthService);
    private router = inject(Router);

    email = '';
    password = '';
    error: WritableSignal<string> = signal('');
    loading: WritableSignal<boolean> = signal(false);

    onSubmit() {
        this.error.set('');
        this.loading.set(true);
        this.auth.register(this.email, this.password).subscribe({
            next: () => {
                this.loading.set(false);
                this.router.navigate(['/dashboard']);
            },
            error: (err) => {
                this.loading.set(false);
                this.error.set(err?.error?.error || 'Registration failed.');
            },
        });
    }
}
