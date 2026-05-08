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
  GeneratePitchRequest,
  GenerationParams,
  PromptTemplate,
  ToneOption,
  LengthOption,
  AngleOption,
  CreatePromptTemplateRequest,
} from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import {
  Plus,
  Sparkles,
  Eye,
  Loader2,
  RotateCcw,
  Save,
  Trash2,
  Columns2,
  X,
} from "lucide-react";

const STATUS_OPTIONS: PitchStatus[] = ["draft", "sent", "opened", "replied"];

const TONE_OPTIONS: { value: ToneOption; label: string }[] = [
  { value: "", label: "Default" },
  { value: "formal", label: "Formal" },
  { value: "conversational", label: "Conversational" },
  { value: "urgent", label: "Urgent" },
  { value: "enthusiastic", label: "Enthusiastic" },
];

const LENGTH_OPTIONS: { value: LengthOption; label: string; desc: string }[] = [
  { value: "", label: "Default", desc: "~200 words" },
  { value: "concise", label: "Concise", desc: "~100 words" },
  { value: "standard", label: "Standard", desc: "~200 words" },
  { value: "detailed", label: "Detailed", desc: "~350 words" },
];

const ANGLE_OPTIONS: { value: AngleOption; label: string }[] = [
  { value: "", label: "Auto" },
  { value: "news_hook", label: "News Hook" },
  { value: "exclusive", label: "Exclusive" },
  { value: "follow_up", label: "Follow-Up" },
  { value: "thought_leadership", label: "Thought Leadership" },
  { value: "event", label: "Event Invite" },
];

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

function parseGenParams(raw: string): GenerationParams | null {
  if (!raw) return null;
  try {
    return JSON.parse(raw) as GenerationParams;
  } catch {
    return null;
  }
}

const toneLabel = (v: string) =>
  TONE_OPTIONS.find((t) => t.value === v)?.label ?? v;
const lengthLabel = (v: string) =>
  LENGTH_OPTIONS.find((l) => l.value === v)?.label ?? v;
const angleLabel = (v: string) =>
  ANGLE_OPTIONS.find((a) => a.value === v)?.label ?? v;

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

  // Generation dialog
  const [genDialogOpen, setGenDialogOpen] = useState(false);
  const [genPitchId, setGenPitchId] = useState<string>("");
  const [genForm, setGenForm] = useState<GeneratePitchRequest>({});
  const [templates, setTemplates] = useState<PromptTemplate[]>([]);
  const [saveTemplateName, setSaveTemplateName] = useState("");
  const [savingTemplate, setSavingTemplate] = useState(false);

  // Refine mode (set when clicking "Refine This" on a version)
  const [refineVersion, setRefineVersion] = useState<PitchVersion | null>(null);

  // Comparison mode
  const [compareMode, setCompareMode] = useState(false);
  const [compareIds, setCompareIds] = useState<string[]>([]);

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

  const fetchTemplates = useCallback(async () => {
    try {
      const data = await api.get<PromptTemplate[]>("/api/prompt-templates");
      setTemplates(data || []);
    } catch {
      // silent
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

  function openGenerateDialog(pitchId: string, refVersion?: PitchVersion) {
    setGenPitchId(pitchId);
    setRefineVersion(refVersion ?? null);
    setGenForm(
      refVersion
        ? {
            reference_version_id: refVersion.id,
            refinement_note: "",
          }
        : {}
    );
    setSaveTemplateName("");
    setGenDialogOpen(true);
    fetchTemplates();
  }

  function loadTemplate(templateId: string) {
    const tmpl = templates.find((t) => t.id === templateId);
    if (!tmpl) return;
    setGenForm((prev) => ({
      ...prev,
      tone: tmpl.tone || undefined,
      length: tmpl.length || undefined,
      angle: tmpl.angle || undefined,
      custom_instructions: tmpl.custom_instructions || undefined,
    }));
    toast.success(`Template "${tmpl.name}" loaded`);
  }

  async function handleSaveTemplate() {
    if (!saveTemplateName.trim()) {
      toast.error("Template name is required");
      return;
    }
    setSavingTemplate(true);
    try {
      const body: CreatePromptTemplateRequest = {
        name: saveTemplateName.trim(),
        tone: genForm.tone,
        length: genForm.length,
        angle: genForm.angle,
        custom_instructions: genForm.custom_instructions,
      };
      await api.post("/api/prompt-templates", body);
      toast.success("Template saved");
      setSaveTemplateName("");
      fetchTemplates();
    } catch (err) {
      toast.error(
        err instanceof ApiError ? err.message : "Failed to save template"
      );
    } finally {
      setSavingTemplate(false);
    }
  }

  async function handleDeleteTemplate(id: string) {
    try {
      await api.del(`/api/prompt-templates/${id}`);
      toast.success("Template deleted");
      fetchTemplates();
    } catch {
      toast.error("Failed to delete template");
    }
  }

  async function handleGenerate(pitchId: string, params?: GeneratePitchRequest) {
    setGenerating(pitchId);
    setGenDialogOpen(false);
    try {
      // Build request body, omitting empty strings.
      const body: Record<string, string> = {};
      if (params?.tone) body.tone = params.tone;
      if (params?.length) body.length = params.length;
      if (params?.angle) body.angle = params.angle;
      if (params?.custom_instructions)
        body.custom_instructions = params.custom_instructions;
      if (params?.reference_version_id)
        body.reference_version_id = params.reference_version_id;
      if (params?.refinement_note)
        body.refinement_note = params.refinement_note;

      const res = await api.post<{ version_number: number }>(
        `/api/pitches/${pitchId}/generate`,
        Object.keys(body).length > 0 ? body : undefined
      );
      const expectedVersion = res.version_number;

      // Poll until the worker saves the version (max 40 × 1.5 s = 60 s)
      let found = false;
      for (let i = 0; i < 40; i++) {
        await new Promise((r) => setTimeout(r, 1500));
        const data = await api.get<PitchVersion[]>(
          `/api/pitches/${pitchId}/versions`
        );
        if (data?.find((v) => v.version_number === expectedVersion)) {
          found = true;
          toast.success(`Version ${expectedVersion} generated`);
          if (viewingPitch?.id === pitchId) setVersions(data);
          break;
        }
      }
      if (!found) toast.error("Generation timed out — try again shortly.");
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
    setCompareMode(false);
    setCompareIds([]);
    fetchVersions(pitch.id);
  }

  function toggleCompare(versionId: string) {
    setCompareIds((prev) => {
      if (prev.includes(versionId)) return prev.filter((id) => id !== versionId);
      if (prev.length >= 2) return [prev[1], versionId];
      return [...prev, versionId];
    });
  }

  const compareVersions = compareIds
    .map((id) => versions.find((v) => v.id === id))
    .filter(Boolean) as PitchVersion[];

  return (
    <div className="flex h-full flex-col gap-6">
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
                    <span className="flex flex-1 text-left text-sm truncate" data-slot="select-value">
                      {createForm.client_id ? (
                        clients.find((c) => c.id === createForm.client_id)?.name ?? "Select client"
                      ) : (
                        <span className="text-muted-foreground">Select client</span>
                      )}
                    </span>
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
                    <span className="flex flex-1 text-left text-sm truncate" data-slot="select-value">
                      {createForm.journalist_id ? (
                        (() => { const j = journalists.find((j) => j.id === createForm.journalist_id); return j ? `${j.name} (${j.outlet?.name || "No outlet"})` : "Select journalist"; })()
                      ) : (
                        <span className="text-muted-foreground">Select journalist</span>
                      )}
                    </span>
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
                    <span className="flex flex-1 text-left text-sm truncate" data-slot="select-value">
                      {createForm.campaign_id ? (
                        campaigns.find((c) => c.id === createForm.campaign_id)?.title ?? "No campaign"
                      ) : (
                        <span className="text-muted-foreground">No campaign</span>
                      )}
                    </span>
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
        <div className="w-48 space-y-1">
          <Label className="text-xs text-muted-foreground">Status</Label>
          <Select
            value={filterStatus}
            onValueChange={(v) => setFilterStatus(v ?? "all")}
          >
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
        <div className="w-64 space-y-1">
          <Label className="text-xs text-muted-foreground">Campaign</Label>
          <Select
            value={filterCampaign}
            onValueChange={(v) => setFilterCampaign(v ?? "all")}
          >
            <SelectTrigger>
              <span className="flex flex-1 text-left text-sm truncate" data-slot="select-value">
                {filterCampaign === "all" ? (
                  <span className="text-muted-foreground">Filter by campaign</span>
                ) : (
                  campaigns.find((c) => c.id === filterCampaign)?.title ?? "Filter by campaign"
                )}
              </span>
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
      <Card className="flex-1 min-h-0">
        <CardHeader>
          <CardTitle>
            Pitches {pitches.length > 0 && `(${pitches.length})`}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex-1 min-h-0 overflow-y-auto">
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
                    <TableCell>{pitch.campaign?.title || "—"}</TableCell>
                    <TableCell>
                      <Select
                        value={pitch.status}
                        onValueChange={(v) => {
                          if (v)
                            handleStatusChange(pitch.id, v as PitchStatus);
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
                          onClick={() => openGenerateDialog(pitch.id)}
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

      {/* ── AI Generation Workbench Dialog ─────────────────────────────────── */}
      <Dialog open={genDialogOpen} onOpenChange={setGenDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Sparkles className="h-5 w-5" />
              {refineVersion
                ? `Refine Version ${refineVersion.version_number}`
                : "AI Generation Workbench"}
            </DialogTitle>
          </DialogHeader>

          <Tabs defaultValue="controls" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="controls">Generation Controls</TabsTrigger>
              <TabsTrigger value="templates">Prompt Templates</TabsTrigger>
            </TabsList>

            {/* ── Controls Tab ──────────────────────────────────────────── */}
            <TabsContent value="controls" className="space-y-4 pt-2">
              {/* Refinement context (when refining a previous version) */}
              {refineVersion && (
                <div className="space-y-2">
                  <Label className="text-sm font-medium">
                    Refining Version {refineVersion.version_number}
                  </Label>
                  <div className="max-h-32 overflow-y-auto whitespace-pre-wrap rounded-md bg-muted p-3 text-xs">
                    {refineVersion.ai_generated_body}
                  </div>
                  <div className="space-y-1">
                    <Label>Refinement Feedback</Label>
                    <Textarea
                      value={genForm.refinement_note || ""}
                      onChange={(e) =>
                        setGenForm({
                          ...genForm,
                          refinement_note: e.target.value,
                        })
                      }
                      rows={2}
                      placeholder="What should be different? e.g. 'Make the opening stronger', 'Add a data point about market size'..."
                    />
                  </div>
                  <Separator />
                </div>
              )}

              <div className="grid grid-cols-3 gap-4">
                {/* Tone */}
                <div className="space-y-1">
                  <Label className="text-xs">Tone</Label>
                  <Select
                    value={genForm.tone || ""}
                    onValueChange={(v) =>
                      setGenForm({ ...genForm, tone: v || undefined })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Default" />
                    </SelectTrigger>
                    <SelectContent>
                      {TONE_OPTIONS.map((t) => (
                        <SelectItem key={t.value || "_default"} value={t.value || "none"}>
                          {t.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Length */}
                <div className="space-y-1">
                  <Label className="text-xs">Length</Label>
                  <Select
                    value={genForm.length || ""}
                    onValueChange={(v) =>
                      setGenForm({ ...genForm, length: v || undefined })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Default" />
                    </SelectTrigger>
                    <SelectContent>
                      {LENGTH_OPTIONS.map((l) => (
                        <SelectItem key={l.value || "_default"} value={l.value || "none"}>
                          {l.label}{" "}
                          <span className="text-muted-foreground">
                            ({l.desc})
                          </span>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Angle */}
                <div className="space-y-1">
                  <Label className="text-xs">Pitch Angle</Label>
                  <Select
                    value={genForm.angle || ""}
                    onValueChange={(v) =>
                      setGenForm({ ...genForm, angle: v || undefined })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Auto" />
                    </SelectTrigger>
                    <SelectContent>
                      {ANGLE_OPTIONS.map((a) => (
                        <SelectItem key={a.value || "_default"} value={a.value || "none"}>
                          {a.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Custom Instructions */}
              <div className="space-y-1">
                <Label className="text-xs">Custom Instructions</Label>
                <Textarea
                  value={genForm.custom_instructions || ""}
                  onChange={(e) =>
                    setGenForm({
                      ...genForm,
                      custom_instructions: e.target.value,
                    })
                  }
                  rows={3}
                  placeholder="Add specific instructions for the AI. e.g. 'Mention the Q4 earnings report', 'Include a question about their recent article on AI'..."
                />
              </div>

              <Button
                className="w-full"
                disabled={generating === genPitchId}
                onClick={() => {
                  const params: GeneratePitchRequest = {};
                  if (genForm.tone && genForm.tone !== "none")
                    params.tone = genForm.tone;
                  if (genForm.length && genForm.length !== "none")
                    params.length = genForm.length;
                  if (genForm.angle && genForm.angle !== "none")
                    params.angle = genForm.angle;
                  if (genForm.custom_instructions)
                    params.custom_instructions = genForm.custom_instructions;
                  if (genForm.reference_version_id)
                    params.reference_version_id = genForm.reference_version_id;
                  if (genForm.refinement_note)
                    params.refinement_note = genForm.refinement_note;
                  handleGenerate(genPitchId, params);
                }}
              >
                {generating === genPitchId ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Sparkles className="mr-2 h-4 w-4" />
                )}
                {refineVersion ? "Refine & Generate" : "Generate Pitch"}
              </Button>
            </TabsContent>

            {/* ── Templates Tab ─────────────────────────────────────────── */}
            <TabsContent value="templates" className="space-y-4 pt-2">
              {/* Load existing template */}
              {templates.length > 0 && (
                <div className="space-y-2">
                  <Label className="text-xs font-medium">
                    Load a Saved Template
                  </Label>
                  <div className="space-y-1">
                    {templates.map((tmpl) => (
                      <div
                        key={tmpl.id}
                        className="flex items-center justify-between rounded-md border px-3 py-2"
                      >
                        <button
                          type="button"
                          className="flex-1 text-left text-sm hover:underline"
                          onClick={() => loadTemplate(tmpl.id)}
                        >
                          <span className="font-medium">{tmpl.name}</span>
                          <span className="ml-2 text-xs text-muted-foreground">
                            {[
                              tmpl.tone && toneLabel(tmpl.tone),
                              tmpl.length && lengthLabel(tmpl.length),
                              tmpl.angle && angleLabel(tmpl.angle),
                            ]
                              .filter(Boolean)
                              .join(" · ") || "No presets"}
                          </span>
                        </button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0 text-destructive"
                          onClick={() => handleDeleteTemplate(tmpl.id)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <Separator />

              {/* Save current settings as template */}
              <div className="space-y-2">
                <Label className="text-xs font-medium">
                  Save Current Settings as Template
                </Label>
                <div className="flex gap-2">
                  <Input
                    value={saveTemplateName}
                    onChange={(e) => setSaveTemplateName(e.target.value)}
                    placeholder="Template name..."
                    className="flex-1"
                  />
                  <Button
                    variant="outline"
                    disabled={savingTemplate || !saveTemplateName.trim()}
                    onClick={handleSaveTemplate}
                  >
                    {savingTemplate ? (
                      <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="mr-1 h-4 w-4" />
                    )}
                    Save
                  </Button>
                </div>
                {(genForm.tone || genForm.length || genForm.angle || genForm.custom_instructions) && (
                  <div className="flex flex-wrap gap-1 text-xs">
                    {genForm.tone && genForm.tone !== "none" && (
                      <Badge variant="secondary">{toneLabel(genForm.tone)}</Badge>
                    )}
                    {genForm.length && genForm.length !== "none" && (
                      <Badge variant="secondary">
                        {lengthLabel(genForm.length)}
                      </Badge>
                    )}
                    {genForm.angle && genForm.angle !== "none" && (
                      <Badge variant="secondary">
                        {angleLabel(genForm.angle)}
                      </Badge>
                    )}
                    {genForm.custom_instructions && (
                      <Badge variant="outline" className="max-w-xs truncate">
                        {genForm.custom_instructions}
                      </Badge>
                    )}
                  </div>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* ── Versions Dialog ────────────────────────────────────────────────── */}
      <Dialog open={versionsOpen} onOpenChange={setVersionsOpen}>
        <DialogContent className="max-w-[90vw] w-[90vw] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <div className="flex items-center justify-between">
              <DialogTitle>
                Pitch Versions — {viewingPitch?.journalist?.name || ""}
                {viewingPitch?.campaign
                  ? ` / ${viewingPitch.campaign.title}`
                  : ""}
              </DialogTitle>
              <div className="flex items-center gap-2">
                {versions.length >= 2 && (
                  <Button
                    variant={compareMode ? "default" : "outline"}
                    size="sm"
                    onClick={() => {
                      setCompareMode(!compareMode);
                      setCompareIds([]);
                    }}
                  >
                    <Columns2 className="mr-1 h-4 w-4" />
                    {compareMode ? "Exit Compare" : "Compare"}
                  </Button>
                )}
              </div>
            </div>
          </DialogHeader>

          {/* Compare mode: side-by-side */}
          {compareMode && (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Select 2 versions to compare side-by-side.
              </p>
              <div className="flex flex-wrap gap-2">
                {versions.map((v) => (
                  <Button
                    key={v.id}
                    variant={compareIds.includes(v.id) ? "default" : "outline"}
                    size="sm"
                    onClick={() => toggleCompare(v.id)}
                  >
                    v{v.version_number}
                  </Button>
                ))}
              </div>
              {compareVersions.length === 2 && (
                <div className="grid grid-cols-2 gap-4">
                  {compareVersions.map((v) => (
                    <Card key={v.id}>
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-sm">
                            Version {v.version_number}
                          </CardTitle>
                          <span className="text-xs text-muted-foreground">
                            {new Date(v.created_at).toLocaleString()}
                          </span>
                        </div>
                        <VersionParamsBadges params={v.generation_params} versions={versions} />
                      </CardHeader>
                      <CardContent>
                        <div className="whitespace-pre-wrap rounded-md bg-muted p-3 text-sm">
                          {v.ai_generated_body}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Normal mode: stacked versions */}
          {!compareMode && (
            <>
              {versions.length === 0 ? (
                <p className="text-muted-foreground py-4">
                  No versions generated yet. Click &quot;Generate&quot; to
                  create one.
                </p>
              ) : (
                <div className="space-y-4">
                  {versions.map((v) => {
                    const params = parseGenParams(v.generation_params);
                    const refVersionNum = params?.reference_version_id
                      ? versions.find(
                          (rv) => rv.id === params.reference_version_id
                        )?.version_number
                      : null;
                    return (
                      <Card key={v.id}>
                        <CardHeader className="pb-2">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <CardTitle className="text-base">
                                Version {v.version_number}
                              </CardTitle>
                              {refVersionNum && (
                                <Badge variant="secondary" className="text-xs">
                                  <RotateCcw className="mr-1 h-3 w-3" />
                                  Refined from v{refVersionNum}
                                </Badge>
                              )}
                            </div>
                            <span className="text-xs text-muted-foreground">
                              {new Date(v.created_at).toLocaleString()}
                            </span>
                          </div>
                          <VersionParamsBadges params={v.generation_params} versions={versions} />
                        </CardHeader>
                        <CardContent>
                          <div className="whitespace-pre-wrap rounded-md bg-muted p-4 text-sm">
                            {v.ai_generated_body}
                          </div>

                          <div className="mt-3 flex items-center gap-2">
                            {viewingPitch && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() =>
                                  openGenerateDialog(viewingPitch.id, v)
                                }
                                disabled={generating === viewingPitch.id}
                              >
                                <RotateCcw className="mr-1 h-3.5 w-3.5" />
                                Refine This
                              </Button>
                            )}
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
                    );
                  })}
                </div>
              )}
              {viewingPitch && (
                <div className="pt-2">
                  <Button
                    onClick={() => openGenerateDialog(viewingPitch.id)}
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
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ── Sub-component: Generation Params Badges ───────────────────────────────────
function VersionParamsBadges({
  params: raw,
  versions,
}: {
  params: string;
  versions: PitchVersion[];
}) {
  const params = parseGenParams(raw);
  if (!params) return null;

  const hasSomething =
    params.tone || params.length || params.angle || params.custom_instructions;
  if (!hasSomething) return null;

  return (
    <div className="flex flex-wrap gap-1 pt-1">
      {params.tone && (
        <Badge variant="outline" className="text-xs">
          {toneLabel(params.tone)}
        </Badge>
      )}
      {params.length && (
        <Badge variant="outline" className="text-xs">
          {lengthLabel(params.length)}
        </Badge>
      )}
      {params.angle && (
        <Badge variant="outline" className="text-xs">
          {angleLabel(params.angle)}
        </Badge>
      )}
      {params.custom_instructions && (
        <Badge variant="outline" className="text-xs max-w-xs truncate">
          {params.custom_instructions}
        </Badge>
      )}
    </div>
  );
}
