import { useJobs, useDeadJobs, useResubmitJob } from '@/hooks/useJobs';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
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
import { RefreshCcw, ChevronLeft, ChevronRight, AlertCircle, List } from 'lucide-react';
import { motion } from 'framer-motion';

export default function JobList() {
    const location = useLocation();
    const navigate = useNavigate();

    const [searchParams, setSearchParams] = useSearchParams();

    const activeTab = location.pathname === '/jobs/dead' ? 'dead' : 'all';
    const page = parseInt(searchParams.get('page') || '1', 10);
    const limit = 20;
    const offset = (page - 1) * limit;

    const jobsQuery = useJobs(limit, offset);
    const deadJobsQuery = useDeadJobs(limit, offset);
    const resubmitJob = useResubmitJob();

    const { data, isLoading, isError, refetch, isFetching } =
        activeTab === 'all' ? jobsQuery : deadJobsQuery;

    const handleNext = () => {
        if (data && data.meta.current_page < data.meta.total_pages) {
            setSearchParams({ page: (page + 1).toString() });
        }
    };

    const handlePrev = () => {
        if (page > 1) {
            setSearchParams({ page: (page - 1).toString() });
        }
    };

    const handleTabChange = (tab: 'all' | 'dead') => {
        if (tab === activeTab) return;
        navigate(tab === 'dead' ? '/jobs/dead' : '/jobs');
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
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Jobs</h2>
            <p className="text-muted-foreground">
              Manage and monitor distributed jobs
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button variant="outline" size="icon" onClick={() => refetch()}>
              <RefreshCcw
                className={`h-4 w-4 ${isFetching ? "animate-spin" : ""}`}
              />
            </Button>
            <CreateJobModal />
          </div>
        </div>

        <div className="flex items-center space-x-1 p-1 bg-muted/50 rounded-lg w-fit">
          <button
            onClick={() => handleTabChange("all")}
            className={`relative px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === "all"
                ? "text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {activeTab === "all" && (
              <motion.div
                layoutId="activeTab"
                className="absolute inset-0 bg-background shadow-sm rounded-md"
                transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
              />
            )}
            <span className="relative z-10 flex items-center gap-2">
              <List className="w-4 h-4" />
              All Jobs
            </span>
          </button>
          <button
            onClick={() => handleTabChange("dead")}
            className={`relative px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === "dead"
                ? "text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {activeTab === "dead" && (
              <motion.div
                layoutId="activeTab"
                className="absolute inset-0 bg-background shadow-sm rounded-md"
                transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
              />
            )}
            <span className="relative z-10 flex items-center gap-2">
              <AlertCircle className="w-4 h-4" />
              Failure Jobs
            </span>
          </button>
        </div>

        <div className="rounded-md border bg-card h-[calc(100vh-320px)] overflow-auto relative">
          <Table>
            <TableHeader className="sticky top-0 z-10 bg-card">
              <TableRow>
                <TableHead className="w-[100px]">ID</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created At</TableHead>
                <TableHead className="text-right">Retries</TableHead>
                {activeTab === "dead" && (
                  <TableHead className="text-right">Actions</TableHead>
                )}
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
                  <TableCell
                    colSpan={5}
                    className="text-center h-24 text-destructive"
                  >
                    Failed to load jobs.
                  </TableCell>
                </TableRow>
              ) : data?.jobs.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center h-24 text-muted-foreground"
                  >
                    No jobs found.
                  </TableCell>
                </TableRow>
              ) : (
                data?.jobs.map((job) => (
                  <TableRow
                    key={job.id}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() =>
                      navigate(`/jobs/${job.id}`, { state: { job } })
                    }
                  >
                    <TableCell className="font-mono text-xs">
                      {(job.id || "").substring(0, 8)}...
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className="font-mono text-xs font-normal"
                      >
                        {job.type}
                      </Badge>
                    </TableCell>
                    <TableCell>{getStatusBadge(job.status)}</TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {job.created_at
                        ? format(new Date(job.created_at), "MMM d, HH:mm:ss")
                        : "N/A"}
                    </TableCell>
                    <TableCell className="text-right font-mono">
                      {job.retry_count}
                    </TableCell>
                    {activeTab === "dead" && (
                      <TableCell className="text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 border-red-900/30 hover:bg-red-900/10 hover:text-red-500 text-red-500"
                          disabled={resubmitJob.isPending}
                          onClick={(e) => {
                            e.stopPropagation(); 
                            resubmitJob.mutate(job);
                          }}
                        >
                          <RefreshCcw
                            className={`w-3 h-3 mr-1 ${resubmitJob.isPending ? "animate-spin" : ""}`}
                          />
                          Resubmit
                        </Button>
                      </TableCell>
                    )}
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
            disabled={
              !data ||
              data.meta.current_page >= data.meta.total_pages ||
              isLoading
            }
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
}

