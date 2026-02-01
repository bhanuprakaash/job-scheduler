import { useNavigate, useParams } from 'react-router-dom';
import { useJob } from '@/hooks/useJobs';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, RotateCcw } from 'lucide-react';
import { format } from 'date-fns';

export default function JobDetails() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: job, isLoading, error } = useJob(id!);

    if (!id) return null;
    if (isLoading) return <div className="p-8">Loading job details...</div>;
    if (error) return <div className="p-8 text-destructive">Error loading job.</div>;
    if (!job) return <div className="p-8">Job not found.</div>;

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'completed': return <Badge variant="success">Completed</Badge>;
            case 'failed': return <Badge variant="destructive">Failed</Badge>;
            case 'running': return <Badge variant="info">Running</Badge>;
            default: return <Badge variant="secondary">Pending</Badge>;
        }
    };

    return (
        <div className="space-y-6 max-w-4xl mx-auto">
            <div className="flex items-center space-x-4">
                <Button variant="ghost" size="icon" onClick={() => navigate('/jobs')}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Job Details</h2>
                    <p className="text-muted-foreground">{job.id}</p>
                </div>
                <div className="ml-auto">
                    {getStatusBadge(job.status)}
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>Metadata</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid grid-cols-2 gap-2 text-sm">
                            <span className="text-muted-foreground">Type</span>
                            <span className="font-medium font-mono">{job.type}</span>

                            <span className="text-muted-foreground">Created At</span>
                            <span className="font-medium">
                                {format(new Date(job.created_at), 'MMM d, yyyy HH:mm:ss')}
                            </span>

                            <span className="text-muted-foreground">Retry Count</span>
                            <div className="flex items-center space-x-2">
                                <RotateCcw className="h-3 w-3 text-muted-foreground" />
                                <span className="font-medium">{job.retry_count}</span>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {job.status === 'failed' && (
                    <Card className="border-destructive/50 bg-destructive/10">
                        <CardHeader>
                            <CardTitle className="text-destructive">Error Message</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <code className="text-sm text-destructive font-mono whitespace-pre-wrap">
                                {job.error_message || "Unknown Error"}
                            </code>
                        </CardContent>
                    </Card>
                )}
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Payload</CardTitle>
                    <CardDescription>The JSON payload for this job</CardDescription>
                </CardHeader>
                <CardContent>
                    <pre className="bg-muted p-4 rounded-md overflow-x-auto text-sm font-mono">
                        {JSON.stringify(JSON.parse(job.payload || "{}"), null, 2)}
                    </pre>
                </CardContent>
            </Card>
        </div>
    );
}
