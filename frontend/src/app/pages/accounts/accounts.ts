import { Component, inject, OnInit, Signal, signal, WritableSignal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../services/api';
import { PlaidService } from '../../services/plaid';
import { Account } from '../../models/models';

@Component({
    selector: 'app-accounts',
    imports: [CommonModule],
    templateUrl: './accounts.html',
    styleUrl: './accounts.css',
})
export class AccountsPage implements OnInit {
    private api = inject(ApiService);
    private plaid = inject(PlaidService);

    accounts: WritableSignal<Account[]> = signal<Account[]>([]);
    loading: WritableSignal<boolean> = signal(true);
    linking: WritableSignal<boolean> = signal(false);
    linkError: WritableSignal<string> = signal('');

    ngOnInit() {
        this.loadAccounts();
    }

    loadAccounts() {
        this.loading.set(true);
        this.api.getAccounts().subscribe({
            next: (data) => {
                this.loading.set(false);
                this.accounts.set(data ?? []);
            },
            error: () => {
                this.loading.set(false);
                this.accounts.set([]);
            },
        });
    }

    async linkAccount() {
        this.linkError.set('');
        this.linking.set(true);
        try {
            await this.plaid.open();
            this.loadAccounts();
        } catch (err: any) {
            this.linkError.set(err?.message || 'Failed to link account.');
        } finally {
            this.linking.set(false);
        }
    }

    formatCurrency(value: string): string {
        const n = parseFloat(value || '0');
        return n.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
    }
}
