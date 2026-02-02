import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, List } from 'lucide-react';
import { cn } from '@/lib/utils';

export function Layout({ children }: { children: React.ReactNode }) {
    const location = useLocation();

    const navItems = [
        { href: '/', label: 'Dashboard', icon: LayoutDashboard },
        { href: '/jobs', label: 'Jobs', icon: List }
    ];

    return (
        <div className="h-screen overflow-hidden bg-background text-foreground flex">
            {/* Sidebar */}
            <aside className="w-64 border-r bg-card flex flex-col">
                <div className="h-16 flex items-center px-6 border-b">
                    <span className="font-bold text-lg">JobScheduler</span>
                </div>
                <nav className="flex-1 p-4 space-y-2">
                    {navItems.map((item) => {
                        const Icon = item.icon;
                        const isActive = item.href === '/'
                            ? location.pathname === '/'
                            : location.pathname.startsWith(item.href);
                        return (
                            <Link
                                key={item.href}
                                to={item.href}
                                className={cn(
                                    "flex items-center space-x-3 px-3 py-2 rounded-md transition-colors text-sm font-medium",
                                    isActive
                                        ? "bg-primary text-primary-foreground"
                                        : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                                )}
                            >
                                <Icon className="h-4 w-4" />
                                <span>{item.label}</span>
                            </Link>
                        );
                    })}
                </nav>
            </aside>

            {/* Main Content */}
            <main className="flex-1 flex flex-col">
                {/* Header */}
                <header className="h-16 border-b flex items-center justify-between px-6 bg-card/50 backdrop-blur-sm sticky top-0 z-10">
                    <h1 className="text-xl font-semibold">
                        {navItems.find((item) =>
                            item.href === '/'
                                ? location.pathname === '/'
                                : location.pathname.startsWith(item.href)
                        )?.label || 'Overview'}
                    </h1>
                </header>

                {/* Page Content */}
                <div className="flex-1 p-6 overflow-auto">
                    {children}
                </div>
            </main>
        </div>
    );
}
