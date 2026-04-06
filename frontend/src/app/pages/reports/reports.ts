import { CommonModule } from '@angular/common';
import { Component, computed, inject, OnInit, signal, WritableSignal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api';
import { Transaction } from '../../models/models';

type PeriodOption = 3 | 6 | 12;

interface SankeyNode {
    id: string;
    label: string;
    secondaryLabel: string;
    amount: number;
    amountLabel: string;
    x: number;
    y: number;
    width: number;
    height: number;
    color: string;
}

interface SankeyLink {
    id: string;
    sourceId: string;
    targetId: string;
    sourceLabel: string;
    targetLabel: string;
    amount: number;
    amountLabel: string;
    thickness: number;
    path: string;
    color: string;
}

interface CategoryBreakdownItem {
    id: string;
    label: string;
    amount: number;
    amountLabel: string;
    share: number;
    shareLabel: string;
    color: string;
}

interface ReportsViewModel {
    chartHeight: number;
    totalSpend: number;
    totalSpendLabel: string;
    averageMonthlySpendLabel: string;
    includedTransactionCount: number;
    categoryCount: number;
    topCategoryLabel: string;
    topCategoryAmountLabel: string;
    rangeLabel: string;
    activeMonthSummary: string;
    sourceNode: SankeyNode | null;
    categoryNodes: SankeyNode[];
    links: SankeyLink[];
    breakdown: CategoryBreakdownItem[];
    hasData: boolean;
}

const PERIOD_OPTIONS: readonly PeriodOption[] = [3, 6, 12];
const NON_SPENDING_CATEGORIES = new Set(['INCOME', 'TRANSFER_IN', 'TRANSFER_OUT', 'LOAN_DISBURSEMENTS']);
const OTHER_CATEGORY_KEY = 'OTHER_CATEGORIES';
const MAX_VISIBLE_CATEGORIES = 8;
const CURRENCY_FORMATTER = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
});

@Component({
    selector: 'app-reports',
    imports: [CommonModule, FormsModule],
    templateUrl: './reports.html',
    styleUrl: './reports.css',
})
export class ReportsPage implements OnInit {
    private api = inject(ApiService);

    readonly periodOptions = PERIOD_OPTIONS;
    readonly selectedPeriod = signal<PeriodOption>(6);
    readonly transactions: WritableSignal<Transaction[]> = signal<Transaction[]>([]);
    readonly loading = signal(true);
    readonly report = computed(() => this.buildReport(this.transactions(), this.selectedPeriod()));

    ngOnInit() {
        this.loadTransactions();
    }

    setPeriodFromValue(value: PeriodOption | string) {
        const period = Number(value);
        if (period === 3 || period === 6 || period === 12) {
            this.selectedPeriod.set(period);
        }
    }

    private loadTransactions() {
        this.loading.set(true);

        this.api.getTransactions().subscribe({
            next: (transactions) => {
                this.transactions.set(transactions ?? []);
                this.loading.set(false);
            },
            error: () => {
                this.transactions.set([]);
                this.loading.set(false);
            },
        });
    }

    private buildReport(transactions: Transaction[], period: PeriodOption): ReportsViewModel {
        const { startDate, endDate } = this.getDateRange(period);
        const rangeLabel = `${this.formatDate(startDate)} to ${this.formatDate(endDate)}`;
        const activeMonthKeys = new Set<string>();
        const categoryTotals = new Map<string, number>();
        let includedTransactionCount = 0;

        for (const transaction of transactions) {
            if (transaction.pending) {
                continue;
            }

            const amount = this.parseAmount(transaction.amount);
            if (amount <= 0) {
                continue;
            }

            const transactionDate = this.parseDate(transaction.date);
            if (!transactionDate) {
                continue;
            }

            if (transactionDate < startDate || transactionDate > endDate) {
                continue;
            }

            const category = this.normalizeCategory(transaction.personal_finance_category);
            if (NON_SPENDING_CATEGORIES.has(category)) {
                continue;
            }

            const monthKey = this.getMonthKey(transactionDate);
            activeMonthKeys.add(monthKey);
            categoryTotals.set(category, (categoryTotals.get(category) ?? 0) + amount);
            includedTransactionCount += 1;
        }

        const totalSpend = [...categoryTotals.values()].reduce((sum, amount) => sum + amount, 0);
        if (totalSpend <= 0) {
            return {
                chartHeight: 420,
                totalSpend: 0,
                totalSpendLabel: CURRENCY_FORMATTER.format(0),
                averageMonthlySpendLabel: CURRENCY_FORMATTER.format(0),
                includedTransactionCount: 0,
                categoryCount: 0,
                topCategoryLabel: 'No spending',
                topCategoryAmountLabel: CURRENCY_FORMATTER.format(0),
                rangeLabel,
                activeMonthSummary: `No cleared spending found in the past ${period} months.`,
                sourceNode: null,
                categoryNodes: [],
                links: [],
                breakdown: [],
                hasData: false,
            };
        }

        const sortedCategories = [...categoryTotals.entries()].sort((a, b) => b[1] - a[1]);
        const topCategoryEntries = sortedCategories.slice(0, MAX_VISIBLE_CATEGORIES);
        const overflowEntries = sortedCategories.slice(MAX_VISIBLE_CATEGORIES);
        const visibleCategoryTotals = new Map<string, number>(topCategoryEntries);
        const overflowAmount = overflowEntries.reduce((sum, [, amount]) => sum + amount, 0);

        if (overflowAmount > 0) {
            visibleCategoryTotals.set(OTHER_CATEGORY_KEY, overflowAmount);
        }

        const orderedCategoryEntries = [...visibleCategoryTotals.entries()].sort((a, b) => b[1] - a[1]);
        const minCategoryLabelHeight = 48;
        const sourceHeight = Math.max(260, Math.min(560, orderedCategoryEntries.length * 64));
        const sourceScale = sourceHeight / totalSpend;
        const rightGap = orderedCategoryEntries.length > 6 ? 12 : 16;
        const categoryVisualHeight = orderedCategoryEntries.reduce(
            (height, [, amount]) => height + Math.max(amount * sourceScale, minCategoryLabelHeight),
            rightGap * Math.max(orderedCategoryEntries.length - 1, 0)
        );
        const chartHeight = Math.max(420, Math.ceil(Math.max(sourceHeight, categoryVisualHeight) + 96));
        const nodeWidth = 24;
        const leftX = 230;
        const rightX = 700;
        const topPadding = 28;
        const scale = sourceScale;
        const categoryNodes: SankeyNode[] = [];
        const categoryNodeMap = new Map<string, SankeyNode>();
        const categoryCursor = new Map<string, number>();
        const sourceNode: SankeyNode = {
            id: 'timeframe:selected',
            label: `Past ${period} Months`,
            secondaryLabel: rangeLabel,
            amount: totalSpend,
            amountLabel: CURRENCY_FORMATTER.format(totalSpend),
            x: leftX,
            y: topPadding,
            width: nodeWidth,
            height: sourceHeight,
            color: 'rgba(99, 102, 241, 0.9)',
        };

        let currentCategoryY = topPadding;
        orderedCategoryEntries.forEach(([categoryKey, amount], index) => {
            const color = this.getCategoryColor(index);
            const node: SankeyNode = {
                id: `category:${categoryKey}`,
                label: this.formatCategory(categoryKey),
                secondaryLabel: `${this.getShare(amount, totalSpend).toFixed(1)}% of selected spending`,
                amount,
                amountLabel: CURRENCY_FORMATTER.format(amount),
                x: rightX,
                y: currentCategoryY,
                width: nodeWidth,
                height: amount * scale,
                color,
            };

            categoryNodes.push(node);
            categoryNodeMap.set(categoryKey, node);
            categoryCursor.set(categoryKey, node.y);
            currentCategoryY += Math.max(node.height, minCategoryLabelHeight) + rightGap;
        });

        const links: SankeyLink[] = [];
        let sourceCursor = sourceNode.y;

        for (const [categoryKey, amount] of orderedCategoryEntries) {
            const categoryNode = categoryNodeMap.get(categoryKey);
            const targetY = categoryCursor.get(categoryKey) ?? categoryNode?.y ?? 0;
            const thickness = amount * scale;

            if (!categoryNode || thickness <= 0) {
                continue;
            }

            const path = this.buildLinkPath(
                sourceNode.x + sourceNode.width,
                sourceCursor + thickness / 2,
                categoryNode.x,
                targetY + thickness / 2
            );

            links.push({
                id: `selected:${categoryKey}`,
                sourceId: sourceNode.id,
                targetId: categoryNode.id,
                sourceLabel: sourceNode.label,
                targetLabel: categoryNode.label,
                amount,
                amountLabel: CURRENCY_FORMATTER.format(amount),
                thickness,
                path,
                color: this.withAlpha(categoryNode.color, 0.4),
            });

            sourceCursor += thickness;
            categoryCursor.set(categoryKey, targetY + thickness);
        }

        const breakdown = categoryNodes.map((node) => ({
            id: node.id,
            label: node.label,
            amount: node.amount,
            amountLabel: node.amountLabel,
            share: this.getShare(node.amount, totalSpend),
            shareLabel: `${this.getShare(node.amount, totalSpend).toFixed(1)}%`,
            color: node.color,
        }));

        return {
            chartHeight,
            totalSpend,
            totalSpendLabel: CURRENCY_FORMATTER.format(totalSpend),
            averageMonthlySpendLabel: CURRENCY_FORMATTER.format(totalSpend / period),
            includedTransactionCount,
            categoryCount: breakdown.length,
            topCategoryLabel: breakdown[0]?.label ?? 'No spending',
            topCategoryAmountLabel: breakdown[0]?.amountLabel ?? CURRENCY_FORMATTER.format(0),
            rangeLabel,
            activeMonthSummary: `Cleared spending across ${activeMonthKeys.size} calendar ${activeMonthKeys.size === 1 ? 'month' : 'months'}.`,
            sourceNode,
            categoryNodes,
            links,
            breakdown,
            hasData: true,
        };
    }

    private buildLinkPath(startX: number, startY: number, endX: number, endY: number): string {
        const curvature = (endX - startX) * 0.42;
        return [
            `M ${startX} ${startY}`,
            `C ${startX + curvature} ${startY}, ${endX - curvature} ${endY}, ${endX} ${endY}`,
        ].join(' ');
    }

    private getDateRange(period: PeriodOption): { startDate: Date; endDate: Date } {
        const endDate = this.startOfDay(new Date());
        const startDate = this.startOfDay(new Date(endDate));
        startDate.setMonth(startDate.getMonth() - period);

        return { startDate, endDate };
    }

    private startOfDay(date: Date): Date {
        return new Date(date.getFullYear(), date.getMonth(), date.getDate());
    }

    private formatDate(date: Date): string {
        return date.toLocaleDateString('en-US', {
            month: 'long',
            day: 'numeric',
            year: 'numeric',
        });
    }

    private parseDate(value: string): Date | null {
        const parsed = new Date(`${value}T00:00:00`);
        return Number.isNaN(parsed.getTime()) ? null : parsed;
    }

    private getMonthKey(date: Date): string {
        const year = date.getFullYear();
        const month = `${date.getMonth() + 1}`.padStart(2, '0');
        return `${year}-${month}`;
    }

    private parseAmount(value: string): number {
        const amount = parseFloat(value || '0');
        return Number.isFinite(amount) ? amount : 0;
    }

    private normalizeCategory(category?: string): string {
        const normalized = (category ?? 'OTHER').trim().toUpperCase();
        return normalized || 'OTHER';
    }

    private formatCategory(category: string): string {
        if (category === OTHER_CATEGORY_KEY) {
            return 'Other Categories';
        }

        return category
            .toLowerCase()
            .replace(/_/g, ' ')
            .replace(/\b\w/g, (character) => character.toUpperCase());
    }

    private getCategoryColor(index: number): string {
        const hue = (index * 47 + 186) % 360;
        return `hsl(${hue} 68% 58%)`;
    }

    private withAlpha(color: string, alpha: number): string {
        if (!color.startsWith('hsl(') || !color.endsWith(')')) {
            return color;
        }

        const colorContent = color.slice(4, -1);
        return `hsl(${colorContent} / ${alpha})`;
    }

    private getShare(amount: number, total: number): number {
        if (total <= 0) {
            return 0;
        }

        return (amount / total) * 100;
    }
}
