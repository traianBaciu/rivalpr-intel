"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import type {
  Pitch,
  PitchVersion,
  Client,
  Campaign,
  Journalist,
  CreatePitchRequest,
  PitchStatus,
  PaginatedResponse,
} from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";
import { Plus, Sparkles, Eye, Loader2 } from "lucide-react";

const STATUS_OPTIONS: PitchStatus[] = ["draft", "sent", "opened", "replied"];

const statusVariant = (s: string) => {
  switch (s) {
    case "replied":
      return "default" as const;
    case "opened":
      return "secondary" as const;
    case "sent":
      return "outline" as const;
    default:
      return "outline" as const;
  }
};

export default function PitchesPage() {
  const [pitches, setPitches] = useState<Pitch[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [journalists, setJournalists] = useState<Journalist[]>([]);
  const [loading, setLoading] = useState(true);

  // Filters
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const [filterCampaign, setFilterCampaign] = useState<string>("all");

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreatePitchRequest>({
    client_id: "",
    journalist_id: "",
    campaign_id: "",
    context_brief: "",
  });

  // Version viewer
  const [viewingPitch, setViewingPitch] = useState<Pitch | null>(null);
  const [versions, setVersions] = useState<PitchVersion[]>([]);
  const [versionsOpen, setVersionsOpen] = useState(false);
  const [generating, setGenerating] = useState<string | null>(null);

  const fetchPitches = useCallback(async () => {
    try {
      let path = "/api/pitches";
      const params: string[] = [];
      if (filterStatus !== "all") params.push(`status=${filterStatus}`);
      if (filterCampaign !== "all")
        params.push(`campaign_id=${filterCampaign}`);
      if (params.length) path += "?" + params.join("&");
      const data = await api.get<Pitch[]>(path);
      setPitches(data || []);
    } catch {
      toast.error("Failed to load pitches");
    }
  }, [filterStatus, filterCampaign]);

  const fetchLookups = useCallback(async () => {
    try {
      const [cl, ca, jo] = await Promise.all([
        api.get<Client[]>("/api/clients"),
        api.get<Campaign[]>("/api/campaigns"),
        api.get<PaginatedResponse<Journalist>>("/api/journalists?limit=100"),
      ]);
      setClients(cl || []);
      setCampaigns(ca || []);
      setJournalists(jo?.data || []);
    } catch {
      // handled
    }
  }, []);

  useEffect(() => {
    Promise.all([fetchPitches(), fetchLookups()]).finally(() =>
      setLoading(false)
    );
  }, [fetchPitches, fetchLookups]);

  useEffect(() => {
    if (!loading) fetchPitches();
  }, [filterStatus, filterCampaign, fetchPitches, loading]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    try {
      const body: Record<string, string> = {
        client_id: createForm.client_id,
        journalist_id: createForm.journalist_id,
      };
      if (createForm.campaign_id) body.campaign_id = createForm.campaign_id;
      if (createForm.context_brief)
        body.context_brief = createForm.context_brief;
      await api.post("/api/pitches", body);
      toast.success("Pitch created");
      setCreateOpen(false);
      setCreateForm({
        client_id: "",
        journalist_id: "",
        campaign_id: "",
        context_brief: "",
      });
      fetchPitches();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Creation failed");
    }
  }

  async function handleStatusChange(pitchId: string, status: PitchStatus) {
    try {
      await api.put(`/api/pitches/${pitchId}`, { status });
      toast.success(`Status updated to ${status}`);
      fetchPitches();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Update failed");
    }
  }

  async function handleGenerate(pitchId: string) {
    setGenerating(pitchId);
    try {
      const version = await api.post<PitchVersion>(
        `/api/pitches/${pitchId}/generate`
      );
      toast.success(`Version ${version.version_number} generated`);
      // If viewing this pitch, refresh versions
      if (viewingPitch?.id === pitchId) {
        fetchVersions(pitchId);
      }
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Generation failed");
    } finally {
      setGenerating(null);
    }
  }

  async function fetchVersions(pitchId: string) {
    try {
      const data = await api.get<PitchVersion[]>(
        `/api/pitches/${pitchId}/versions`
      );
      setVersions(data || []);
    } catch {
      toast.error("Failed to load versions");
    }
  }

  function openVersions(pitch: Pitch) {
    setViewingPitch(pitch);
    setVersionsOpen(true);
    fetchVersions(pitch.id);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Pitches</h1>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger render={<Button />}>
            <Plus className="mr-2 h-4 w-4" /> New Pitch
          </DialogTrigger>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>Create Pitch</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div className="space-y-2">
                <Label>Client</Label>
                <Select
                  value={createForm.client_id}
                  onValueChange={(v) =>
                    setCreateForm({ ...createForm, client_id: v ?? "" })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select client" />
                  </SelectTrigger>
                  <SelectContent>
                    {clients.map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Journalist</Label>
                <Select
                  value={createForm.journalist_id}
                  onValueChange={(v) =>
                    setCreateForm({ ...createForm, journalist_id: v ?? "" })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select journalist" />
                  </SelectTrigger>
                  <SelectContent>
                    {journalists.map((j) => (
                      <SelectItem key={j.id} value={j.id}>
                        {j.name} ({j.outlet?.name || "No outlet"})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Campaign (optional)</Label>
                <Select
                  value={createForm.campaign_id || "none"}
                  onValueChange={(v) =>
                    setCreateForm({
                      ...createForm,
                      campaign_id: v === "none" || !v ? "" : v,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="No campaign" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">No campaign</SelectItem>
                    {campaigns.map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.title}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Context Brief</Label>
                <Textarea
                  value={createForm.context_brief || ""}
                  onChange={(e) =>
                    setCreateForm({
                      ...createForm,
                      context_brief: e.target.value,
                    })
                  }
                  rows={3}
                  placeholder="Additional context for AI generation..."
                />
              </div>
              <Button type="submit" className="w-full">
                Create Pitch
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="w-48">
          <Select value={filterStatus} onValueChange={(v) => setFilterStatus(v ?? "all")}>
            <SelectTrigger>
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              {STATUS_OPTIONS.map((s) => (
                <SelectItem key={s} value={s}>
                  {s.charAt(0).toUpperCase() + s.slice(1)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="w-64">
          <Select value={filterCampaign} onValueChange={(v) => setFilterCampaign(v ?? "all")}>
            <SelectTrigger>
              <SelectValue placeholder="Filter by campaign" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Campaigns</SelectItem>
              {campaigns.map((c) => (
                <SelectItem key={c.id} value={c.id}>
                  {c.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Pitches table */}
      <Card>
        <CardHeader>
          <CardTitle>
            Pitches {pitches.length > 0 && `(${pitches.length})`}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-muted-foreground">Loading...</p>
          ) : pitches.length === 0 ? (
            <p className="text-muted-foreground">
              No pitches found. Create one above.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Journalist</TableHead>
                  <TableHead>Client</TableHead>
                  <TableHead>Campaign</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="w-48">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pitches.map((pitch) => (
                  <TableRow key={pitch.id}>
                    <TableCell className="font-medium">
                      {pitch.journalist?.name || "—"}
                    </TableCell>
                    <TableCell>{pitch.client?.name || "—"}</TableCell>
                    <TableCell>
                      {pitch.campaign?.title || "—"}
                    </TableCell>
                    <TableCell>
                      <Select
                        value={pitch.status}
                        onValueChange={(v) => {
                          if (v) handleStatusChange(pitch.id, v as PitchStatus);
                        }}
                      >
                        <SelectTrigger className="w-28">
                          <Badge variant={statusVariant(pitch.status)}>
                            {pitch.status}
                          </Badge>
                        </SelectTrigger>
                        <SelectContent>
                          {STATUS_OPTIONS.map((s) => (
                            <SelectItem key={s} value={s}>
                              {s}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell>
                      {new Date(pitch.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={generating === pitch.id}
                          onClick={() => handleGenerate(pitch.id)}
                        >
                          {generating === pitch.id ? (
                            <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                          ) : (
                            <Sparkles className="mr-1 h-4 w-4" />
                          )}
                          Generate
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openVersions(pitch)}
                        >
                          <Eye className="mr-1 h-4 w-4" />
                          Versions
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Versions dialog */}
      <Dialog open={versionsOpen} onOpenChange={setVersionsOpen}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              Pitch Versions — {viewingPitch?.journalist?.name || ""}
              {viewingPitch?.campaign
                ? ` / ${viewingPitch.campaign.title}`
                : ""}
            </DialogTitle>
          </DialogHeader>
          {versions.length === 0 ? (
            <p className="text-muted-foreground py-4">
              No versions generated yet. Click &quot;Generate&quot; to create
              one.
            </p>
          ) : (
            <div className="space-y-4">
              {versions.map((v) => (
                <Card key={v.id}>
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">
                        Version {v.version_number}
                      </CardTitle>
                      <span className="text-xs text-muted-foreground">
                        {new Date(v.created_at).toLocaleString()}
                      </span>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="whitespace-pre-wrap rounded-md bg-muted p-4 text-sm">
                      {v.ai_generated_body}
                    </div>
                    {v.prompt_snapshot && (
                      <>
                        <Separator className="my-3" />
                        <details>
                          <summary className="cursor-pointer text-xs text-muted-foreground">
                            View prompt snapshot
                          </summary>
                          <pre className="mt-2 whitespace-pre-wrap rounded-md bg-muted/50 p-3 text-xs">
                            {v.prompt_snapshot}
                          </pre>
                        </details>
                      </>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
          {viewingPitch && (
            <div className="pt-2">
              <Button
                onClick={() => handleGenerate(viewingPitch.id)}
                disabled={generating === viewingPitch.id}
              >
                {generating === viewingPitch.id ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Sparkles className="mr-2 h-4 w-4" />
                )}
                Generate New Version
              </Button>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
