import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import {
  CheckCircle2,
  FileIcon,
  Trash2,
  UploadCloud,
  XCircle,
} from "lucide-react";
import { useCallback, useEffect, useReducer, useState } from "react";
import { useDropzone } from "react-dropzone";
import { toast } from "sonner";
import { useUploadFile } from "../api/useUploadFile";

const MAX_FILE_SIZE = 500 * 1024 * 1024; // 500mb

enum FileUploadStatus {
  QUEUED = "queued",
  PENDING = "pending",
  SUCCESS = "success",
  ERROR = "error",
}

interface FileUpload {
  id: string;
  file: File;
  status: FileUploadStatus;
}

interface UploadState {
  files: FileUpload[];
}

type FileAction =
  | { type: "ADD_FILES"; files: File[] }
  | { type: "UPDATE_FILE_STATUS"; id: string; status: FileUploadStatus }
  | { type: "REMOVE_FILE"; id: string }
  | { type: "RESET" };

function reducer(state: UploadState, action: FileAction): UploadState {
  switch (action.type) {
    case "ADD_FILES": {
      const newFiles = action.files.map((file) => ({
        id: Math.random().toString(36).substring(2, 9),
        file,
        status: FileUploadStatus.QUEUED,
      }));
      return {
        ...state,
        files: [...state.files, ...newFiles],
      };
    }
    case "UPDATE_FILE_STATUS":
      return {
        ...state,
        files: state.files.map((file) =>
          file.id === action.id ? { ...file, status: action.status } : file,
        ),
      };
    case "REMOVE_FILE":
      return {
        ...state,
        files: state.files.filter((file) => file.id !== action.id),
      };
    case "RESET":
      return { files: [] };
    default:
      return state;
  }
}

export function UploadDialog() {
  const { mutateAsync } = useUploadFile();
  const [isOpen, setIsOpen] = useState(false);

  const [uploadState, dispatch] = useReducer(reducer, {
    files: [],
  });

  const onDrop = useCallback((acceptedFiles: File[]) => {
    const validFiles = acceptedFiles.filter(
      (file) => file.size > 0 && file.size <= MAX_FILE_SIZE,
    );
    if (validFiles.length > 0) {
      dispatch({ type: "ADD_FILES", files: validFiles });
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive, open } = useDropzone({
    maxSize: MAX_FILE_SIZE,
    onDrop,
    noClick: true,
    noKeyboard: true,
  });

  useEffect(() => {
    const handlePaste = (event: ClipboardEvent) => {
      if (event.clipboardData && event.clipboardData.files.length > 0) {
        const files = Array.from(event.clipboardData.files);
        onDrop(files);
      }
    };

    if (isOpen) {
      window.addEventListener("paste", handlePaste);
    }

    return () => {
      window.removeEventListener("paste", handlePaste);
    };
  }, [isOpen, onDrop]);

  const handleUpload = async () => {
    const uploadPromises = uploadState.files.map(async (file) => {
      if (file.status !== FileUploadStatus.QUEUED) return "skip" as const;

      dispatch({
        type: "UPDATE_FILE_STATUS",
        id: file.id,
        status: FileUploadStatus.PENDING,
      });

      try {
        await mutateAsync(file.file);
        dispatch({
          type: "UPDATE_FILE_STATUS",
          id: file.id,
          status: FileUploadStatus.SUCCESS,
        });
        return "success" as const;
      } catch {
        dispatch({
          type: "UPDATE_FILE_STATUS",
          id: file.id,
          status: FileUploadStatus.ERROR,
        });
        return "error" as const;
      }
    });

    const results = await Promise.all(uploadPromises);
    const succeeded = results.filter((r) => r === "success").length;
    const failed = results.filter((r) => r === "error").length;
    if (succeeded > 0) {
      toast.success(`Uploaded ${succeeded} file${succeeded > 1 ? "s" : ""}`);
    }
    if (failed > 0) {
      toast.error(`${failed} file${failed > 1 ? "s" : ""} failed to upload`);
    }
  };

  const handleDialogOpenChange = (open: boolean) => {
    setIsOpen(open);
    if (!open) {
      dispatch({ type: "RESET" });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleDialogOpenChange}>
      <DialogTrigger asChild>
        <Button>Upload files</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Upload files</DialogTitle>
        </DialogHeader>

        <div
          {...getRootProps()}
          className={cn(
            "cursor-pointer rounded-lg border-2 border-dashed border-border p-8 text-center transition-colors duration-150",
            isDragActive
              ? "border-primary bg-primary/5"
              : "hover:border-border-strong hover:bg-muted/40",
          )}
        >
          <input {...getInputProps()} />
          <div className="flex flex-col items-center justify-center gap-4">
            <div
              className={cn(
                "flex size-14 items-center justify-center rounded-full bg-muted text-muted-foreground transition-colors",
                isDragActive && "bg-primary/10 text-primary",
              )}
            >
              <UploadCloud className="size-7" />
            </div>
            <div className="space-y-1">
              <h3 className="text-sm font-semibold">
                {isDragActive
                  ? "Drop to upload"
                  : "Drag and drop, or paste from clipboard"}
              </h3>
              <p className="text-xs text-muted-foreground">
                Images, video or audio — up to 500MB each
              </p>
            </div>
            <Button type="button" onClick={open} variant="secondary" size="sm">
              Select files
            </Button>
          </div>
        </div>

        {uploadState.files.length > 0 && (
          <div className="space-y-2 mt-4 max-h-60 overflow-y-auto pr-2">
            <h4 className="text-sm font-medium">Files to upload</h4>
            <ul className="divide-y divide-border rounded-md border">
              {uploadState.files.map((fileUpload) => (
                <li
                  key={fileUpload.id}
                  className="flex items-center justify-between p-3"
                >
                  <div className="flex items-center gap-3 overflow-hidden">
                    <div className="p-2 bg-muted rounded-md shrink-0">
                      <FileIcon className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="text-sm overflow-hidden">
                      <p
                        className="font-medium truncate max-w-[200px]"
                        title={fileUpload.file.name}
                      >
                        {fileUpload.file.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {(fileUpload.file.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    {fileUpload.status === FileUploadStatus.QUEUED && (
                      <span className="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground">
                        Queued
                      </span>
                    )}
                    {fileUpload.status === FileUploadStatus.PENDING && (
                      <div className="flex items-center gap-2 rounded-full border border-warning/20 bg-warning/10 px-2 py-1 text-xs text-warning">
                        <div className="size-3 animate-spin rounded-full border-2 border-warning border-t-transparent" />
                        Uploading…
                      </div>
                    )}
                    {fileUpload.status === FileUploadStatus.SUCCESS && (
                      <span className="flex items-center gap-1 rounded-full border border-success/20 bg-success/10 px-2 py-1 text-xs text-success">
                        <CheckCircle2 className="size-3" />
                        Uploaded
                      </span>
                    )}
                    {fileUpload.status === FileUploadStatus.ERROR && (
                      <span className="flex items-center gap-1 rounded-full border border-destructive/20 bg-destructive/10 px-2 py-1 text-xs text-destructive">
                        <XCircle className="size-3" />
                        Failed
                      </span>
                    )}
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      onClick={() =>
                        dispatch({ type: "REMOVE_FILE", id: fileUpload.id })
                      }
                      disabled={fileUpload.status === FileUploadStatus.PENDING}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => setIsOpen(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleUpload}
            disabled={
              uploadState.files.length === 0 ||
              uploadState.files.every(
                (f) => f.status === FileUploadStatus.SUCCESS,
              ) ||
              uploadState.files.some(
                (f) => f.status === FileUploadStatus.PENDING,
              )
            }
          >
            Upload{" "}
            {uploadState.files.length > 0
              ? `(${uploadState.files.length})`
              : ""}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
