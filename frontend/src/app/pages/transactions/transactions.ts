import { Component, inject, OnInit, signal, WritableSignal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../services/api';
import { Transaction } from '../../models/models';

@Component({
    selector: 'app-transactions',
    imports: [CommonModule],
    templateUrl: './transactions.html',
    styleUrl: './transactions.css',
})
export class TransactionsPage implements OnInit {
    private api = inject(ApiService);

    transactions: WritableSignal<Transaction[]> = signal<Transaction[]>([]);
    loading: WritableSignal<boolean> = signal(true);

    ngOnInit() {
        this.api.getTransactions().subscribe({
            next: (data) => {
                this.loading.set(false);
                this.transactions.set(data ?? []);
            },
            error: () => {
                this.loading.set(false);
                this.transactions.set([]);
            },
        });
    }

    formatCurrency(value: string): string {
        const n = parseFloat(value || '0');
        return n.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
    }

    formatDate(dateStr: string): string {
        return new Date(dateStr).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
        });
    }

    parseFloat(value: string): number {
        return parseFloat(value || '0');
    }

    formatCategory(category: string): string {
        return category
            .toLowerCase()
            .replace(/_/g, ' ')
            .replace(/\b\w/g, (c) => c.toUpperCase());
    }
}
