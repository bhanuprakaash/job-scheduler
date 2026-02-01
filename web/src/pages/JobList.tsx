import { useState } from 'react';
import { useJobs } from '@/hooks/useJobs';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { CreateJobModal } from '@/components/CreateJobModal';
import { RefreshCcw, ChevronLeft, ChevronRight } from 'lucide-react';

export default function JobList() {
    const [page, setPage] = useState(1);
    const limit = 20;
    const offset = (page - 1) * limit;
    const { data, isLoading, isError, refetch, isFetching } = useJobs(limit, offset);
    const navigate = useNavigate();

    const handleNext = () => {
        if (data && data.meta.current_page < data.meta.total_pages) {
            setPage((p) => p + 1);
        }
    };

    const handlePrev = () => {
        if (page > 1) {
            setPage((p) => p - 1);
        }
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'completed': return <Badge variant="success">Completed</Badge>;
            case 'failed': return <Badge variant="destructive">Failed</Badge>;
            case 'running': return <Badge variant="info">Running</Badge>;
            default: return <Badge variant="secondary">Pending</Badge>;
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Jobs</h2>
                    <p className="text-muted-foreground">Manage and monitor distributed jobs</p>
                </div>
                <div className="flex items-center space-x-2">
                    <Button variant="outline" size="icon" onClick={() => refetch()}>
                        <RefreshCcw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
                    </Button>
                    <CreateJobModal />
                </div>
            </div>

            <div className="rounded-md border bg-card h-[calc(100vh-280px)] overflow-auto relative">
                <Table>
                    <TableHeader className="sticky top-0 z-10 bg-card">
                        <TableRow>
                            <TableHead className="w-[100px]">ID</TableHead>
                            <TableHead>Type</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Created At</TableHead>
                            <TableHead className="text-right">Retries</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center h-24">
                                    Loading jobs...
                                </TableCell>
                            </TableRow>
                        ) : isError ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center h-24 text-destructive">
                                    Failed to load jobs.
                                </TableCell>
                            </TableRow>
                        ) : data?.jobs.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                    No jobs found.
                                </TableCell>
                            </TableRow>
                        ) : (
                            data?.jobs.map((job) => (
                                <TableRow
                                    key={job.id}
                                    className="cursor-pointer hover:bg-muted/50"
                                    onClick={() => navigate(`/jobs/${job.id}`)}
                                >
                                    <TableCell className="font-mono text-xs">{job.id.substring(0, 8)}...</TableCell>
                                    <TableCell>
                                        <Badge variant="outline" className="font-mono text-xs font-normal">
                                            {job.type}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{getStatusBadge(job.status)}</TableCell>
                                    <TableCell className="text-muted-foreground text-xs">
                                        {format(new Date(job.created_at), 'MMM d, HH:mm:ss')}
                                    </TableCell>
                                    <TableCell className="text-right font-mono">{job.retry_count}</TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            <div className="flex items-center justify-end space-x-2">
                <div className="text-xs text-muted-foreground">
                    Page {data?.meta.current_page || 1} of {data?.meta.total_pages || 1}
                </div>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={handlePrev}
                    disabled={page <= 1 || isLoading}
                >
                    <ChevronLeft className="h-4 w-4" />
                    Previous
                </Button>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={handleNext}
                    disabled={!data || data.meta.current_page >= data.meta.total_pages || isLoading}
                >
                    Next
                    <ChevronRight className="h-4 w-4" />
                </Button>
            </div>
        </div>
    );
}
