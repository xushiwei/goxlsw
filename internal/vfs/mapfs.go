package vfs

import (
	"io"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"time"
)

// MapFile represents a file's content and metadata in the map file system.
type MapFile struct {
	Content []byte
	ModTime time.Time
}

// GetFileMapFunc is the type for function that returns a map of files.
type GetFileMapFunc func() map[string]MapFile

// MapFS implements [fs.ReadDirFS] using a map of files.
type MapFS struct {
	getFileMap    GetFileMapFunc
	fileMode      fs.FileMode
	dirMode       fs.FileMode
	snapshottedAt time.Time
}

// NewMapFS creates a new map file system.
func NewMapFS(getFileMap GetFileMapFunc) *MapFS {
	return &MapFS{
		getFileMap: getFileMap,
		fileMode:   0o444,
		dirMode:    0o444 | fs.ModeDir,
	}
}

// Snapshot returns a snapshot of the map file system. It returns the same
// instance if it is already a snapshot.
func (mfs *MapFS) Snapshot() *MapFS {
	if !mfs.snapshottedAt.IsZero() {
		return mfs
	}
	fileMap := mfs.getFileMap()
	mapFS := NewMapFS(func() map[string]MapFile {
		return fileMap
	})
	mapFS.snapshottedAt = time.Now()
	return mapFS
}

// SnapshottedAt returns the time when the map file system was snapshotted, or
// zero if it is not a snapshot.
func (mfs *MapFS) SnapshottedAt() time.Time {
	return mfs.snapshottedAt
}

// WithOverlay returns a new [MapFS] that overlays the given files on top of the
// existing files. Files in the overlay take precedence over existing files with
// the same name.
func (mfs *MapFS) WithOverlay(overlay map[string]MapFile) *MapFS {
	getFileMap := mfs.getFileMap
	return NewMapFS(func() map[string]MapFile {
		fileMap := maps.Clone(getFileMap())
		maps.Copy(fileMap, overlay)
		return fileMap
	})
}

// Open implements [fs.ReadDirFS].
func (mfs *MapFS) Open(name string) (fs.File, error) {
	fileMap := mfs.getFileMap()

	name = cleanPath(name)
	if name == "" {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	mf, ok := fileMap[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &file{
		name:    name,
		content: mf.Content,
		mode:    mfs.fileMode,
		modTime: mf.ModTime,
	}, nil
}

// ReadDir implements [fs.ReadDirFS].
func (mfs *MapFS) ReadDir(name string) ([]fs.DirEntry, error) {
	fileMap := mfs.getFileMap()

	name = cleanPath(name)
	if name == "" {
		name = "."
	}
	if name != "." {
		// Check if directory exists by looking for files with this prefix.
		var hasPrefix bool
		for p := range fileMap {
			if strings.HasPrefix(p, name+"/") {
				hasPrefix = true
				break
			}
		}
		if !hasPrefix {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
		}
	}

	dirs := make(map[string]struct{})
	files := make(map[string]struct{})
	prefix := ""
	if name != "." {
		prefix = name + "/"
	}
	for p := range fileMap {
		if !strings.HasPrefix(p, prefix) {
			continue
		}

		relPath := p[len(prefix):]
		parts := strings.Split(relPath, "/")
		if len(parts) == 1 {
			// It's a file in the current directory.
			files[parts[0]] = struct{}{}
		} else if len(parts) > 1 && parts[0] != "" {
			// It's a subdirectory.
			dirs[parts[0]] = struct{}{}
		}
	}

	entries := make([]fs.DirEntry, 0, len(dirs)+len(files))
	for d := range dirs {
		var latestModTime time.Time
		dirPrefix := prefix + d + "/"
		for p, mf := range fileMap {
			if strings.HasPrefix(p, dirPrefix) && mf.ModTime.After(latestModTime) {
				latestModTime = mf.ModTime
			}
		}
		entries = append(entries, &dirEntry{
			name:    d,
			mode:    mfs.dirMode,
			modTime: latestModTime,
			isDir:   true,
		})
	}
	for f := range files {
		mf := fileMap[prefix+f]
		entries = append(entries, &dirEntry{
			name:    f,
			size:    int64(len(mf.Content)),
			mode:    mfs.fileMode,
			modTime: mf.ModTime,
			isDir:   false,
		})
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return entries, nil
}

// Stat implements [fs.StatFS].
func (mfs *MapFS) Stat(name string) (fs.FileInfo, error) {
	if name == "" {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	fileMap := mfs.getFileMap()
	if name == "." {
		var latestModTime time.Time
		for _, mf := range fileMap {
			if mf.ModTime.After(latestModTime) {
				latestModTime = mf.ModTime
			}
		}
		return &fileInfo{
			name:    ".",
			mode:    mfs.dirMode,
			modTime: latestModTime,
			isDir:   true,
		}, nil
	}

	mf, ok := fileMap[name]
	if ok {
		return &fileInfo{
			name:    path.Base(name),
			size:    int64(len(mf.Content)),
			mode:    mfs.fileMode,
			modTime: mf.ModTime,
			isDir:   false,
		}, nil
	}

	// Check if it's a directory by looking for files with this prefix.
	prefix := name + "/"
	hasPrefix := false
	var latestModTime time.Time
	for p, mf := range fileMap {
		if strings.HasPrefix(p, prefix) {
			hasPrefix = true
			if mf.ModTime.After(latestModTime) {
				latestModTime = mf.ModTime
			}
		}
	}

	if hasPrefix {
		return &fileInfo{
			name:    path.Base(name),
			mode:    mfs.dirMode,
			modTime: latestModTime,
			isDir:   true,
		}, nil
	}

	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

// file implements [fs.file].
type file struct {
	name    string
	content []byte
	mode    fs.FileMode
	modTime time.Time
	offset  int64
}

// Stat implements [fs.File].
func (f *file) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name:    path.Base(f.name),
		size:    int64(len(f.content)),
		mode:    f.mode,
		modTime: f.modTime,
		isDir:   false,
	}, nil
}

// Read implements [fs.File].
func (f *file) Read(b []byte) (int, error) {
	if f.offset >= int64(len(f.content)) {
		return 0, io.EOF
	}

	n := copy(b, f.content[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// Close implements [fs.File].
func (f *file) Close() error {
	return nil
}

// fileInfo implements [fs.fileInfo].
type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() any           { return nil }

// dirEntry implements [fs.dirEntry].
type dirEntry struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (de *dirEntry) Name() string      { return de.name }
func (de *dirEntry) IsDir() bool       { return de.isDir }
func (de *dirEntry) Type() fs.FileMode { return de.mode }
func (de *dirEntry) Info() (fs.FileInfo, error) {
	return &fileInfo{
		name:    de.name,
		size:    de.size,
		mode:    de.mode,
		modTime: de.modTime,
		isDir:   de.isDir,
	}, nil
}

// cleanPath returns the cleaned path.
func cleanPath(name string) string {
	return path.Clean("/" + name)[1:]
}
