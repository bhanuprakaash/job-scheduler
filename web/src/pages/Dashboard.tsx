import { useJobStats } from '@/hooks/useJobs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, CheckCircle2, Clock, XCircle, AlertCircle, type LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    PieChart,
    Pie,
    Cell,
} from 'recharts';

export default function Dashboard() {
    const { data: stats, isLoading, error } = useJobStats();

    if (isLoading) return <div className="p-8 flex justify-center text-muted-foreground">Loading dashboard...</div>;
    if (error) return <div className="p-8 text-destructive">Error loading stats.</div>;
    if (!stats) return null;

    const chartData = [
        { name: 'Pending', value: stats.pending_jobs, color: '#eab308' }, // yellow-500
        { name: 'Running', value: stats.running_jobs, color: '#3b82f6' }, // blue-500
        { name: 'Completed', value: stats.completed_jobs, color: '#22c55e' }, // green-500
        { name: 'Failed', value: stats.failed_jobs, color: '#ef4444' },   // red-500
    ];

    const pieData = chartData.filter(d => d.value > 0);

    return (
        <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Top Stats Row */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
                <StatsCard
                    title="Total Jobs"
                    value={stats.total_jobs}
                    icon={Activity}
                    className="col-span-1 lg:col-span-1 border-primary/20 bg-primary/5"
                    iconClassName="text-primary"
                />
                <StatsCard
                    title="Pending"
                    value={stats.pending_jobs}
                    icon={Clock}
                    className="bg-card/50"
                    iconClassName="text-yellow-500"
                />
                <StatsCard
                    title="Running"
                    value={stats.running_jobs}
                    icon={AlertCircle}
                    className="bg-card/50"
                    iconClassName="text-blue-500"
                />
                <StatsCard
                    title="Completed"
                    value={stats.completed_jobs}
                    icon={CheckCircle2}
                    className="bg-green-950/10 border-green-900/20"
                    valueClassName="text-green-500"
                    iconClassName="text-green-500"
                />
                <StatsCard
                    title="Failed"
                    value={stats.failed_jobs}
                    icon={XCircle}
                    className="bg-red-950/10 border-red-900/20"
                    valueClassName="text-red-500"
                    iconClassName="text-red-500"
                />
            </div>

            {/* Charts Row */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
                {/* Bar Chart: Job Status Distribution */}
                <Card className="col-span-4 bg-card/50 backdrop-blur-sm border-muted/20">
                    <CardHeader>
                        <CardTitle>Job Status Distribution</CardTitle>
                    </CardHeader>
                    <CardContent className="pl-2">
                        <div className="relative h-[300px] w-full min-w-0">
                            <div className="absolute inset-0">
                                <ResponsiveContainer width="100%" height="100%">
                                    <BarChart data={chartData}>
                                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.5} vertical={false} />
                                        <XAxis
                                            dataKey="name"
                                            stroke="#94a3b8"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                        />
                                        <YAxis
                                            stroke="#94a3b8"
                                            fontSize={12}
                                            tickLine={false}
                                            axisLine={false}
                                            tickFormatter={(value) => `${value}`}
                                        />
                                        <Tooltip
                                            cursor={{ fill: 'transparent' }}
                                            contentStyle={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#f8fafc' }}
                                            itemStyle={{ color: '#f8fafc' }}
                                        />
                                        <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                                            {chartData.map((entry, index) => (
                                                <Cell key={`cell-${index}`} fill={entry.color} />
                                            ))}
                                        </Bar>
                                    </BarChart>
                                </ResponsiveContainer>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Pie Chart: Completion Rate */}
                <Card className="col-span-3 bg-card/50 backdrop-blur-sm border-muted/20">
                    <CardHeader>
                        <CardTitle>Completion Ratio</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="relative h-[300px] w-full min-w-0">
                            {stats.total_jobs === 0 ? (
                                <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground">
                                    <Activity className="h-10 w-10 mb-2 opacity-50" />
                                    <span>No data available</span>
                                </div>
                            ) : (
                                <div className="absolute inset-0">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <PieChart>
                                            <Pie
                                                data={pieData}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={60}
                                                outerRadius={80}
                                                paddingAngle={5}
                                                dataKey="value"
                                                stroke="none"
                                            >
                                                {pieData.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip
                                                contentStyle={{ backgroundColor: '#1e293b', borderColor: '#334155', color: '#f8fafc', borderRadius: '6px' }}
                                                itemStyle={{ color: '#f8fafc' }}
                                            />
                                        </PieChart>
                                    </ResponsiveContainer>
                                </div>
                            )}
                        </div>
                        {stats.total_jobs > 0 && (
                            <div className="flex justify-center gap-4 mt-[-20px]">
                                {chartData.map((item) => (
                                    <div key={item.name} className="flex items-center gap-2 text-xs">
                                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                                        <span className="text-muted-foreground">{item.name}</span>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

function StatsCard({
    title,
    value,
    icon: Icon,
    className,
    valueClassName,
    iconClassName
}: {
    title: string;
    value: number;
    icon: LucideIcon;
    className?: string;
    valueClassName?: string;
    iconClassName?: string;
}) {
    return (
        <Card className={cn(className)}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                    {title}
                </CardTitle>
                <Icon className={cn("h-4 w-4 text-muted-foreground", iconClassName)} />
            </CardHeader>
            <CardContent>
                <div className={cn("text-2xl font-bold", valueClassName)}>{value}</div>
            </CardContent>
        </Card>
    );
}
