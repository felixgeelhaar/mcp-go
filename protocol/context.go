package protocol

import "context"

// requestMetaKey is the context key for request metadata.
type requestMetaKey struct{}

// RequestMeta holds metadata associated with a request.
// This is typically used to pass HTTP headers or other transport-level
// information to middleware and handlers.
type RequestMeta map[string]string

// ContextWithRequestMeta returns a new context with the request metadata attached.
func ContextWithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey{}, meta)
}

// RequestMetaFromContext returns the request metadata from the context.
// Returns nil if no metadata is present.
func RequestMetaFromContext(ctx context.Context) RequestMeta {
	if meta, ok := ctx.Value(requestMetaKey{}).(RequestMeta); ok {
		return meta
	}
	return nil
}

// GetRequestMeta returns a specific metadata value from the context.
// Returns empty string if the key is not found or no metadata is present.
func GetRequestMeta(ctx context.Context, key string) string {
	meta := RequestMetaFromContext(ctx)
	if meta == nil {
		return ""
	}
	return meta[key]
}

// protocolVersionKey is the context key for the negotiated (or per-request)
// MCP protocol version.
type protocolVersionKey struct{}

// ContextWithProtocolVersion returns a context carrying the protocol version
// in force for this request (initialize-negotiated, or `_meta` on the modern
// path).
func ContextWithProtocolVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, protocolVersionKey{}, version)
}

// ProtocolVersionFromContext returns the protocol version attached to ctx, or
// empty string if none was set.
func ProtocolVersionFromContext(ctx context.Context) string {
	v, _ := ctx.Value(protocolVersionKey{}).(string)
	return v
}

// SetRequestMeta sets a metadata value in the context.
// If no metadata exists, a new map is created.
func SetRequestMeta(ctx context.Context, key, value string) context.Context {
	meta := RequestMetaFromContext(ctx)
	if meta == nil {
		meta = make(RequestMeta)
	} else {
		// Create a copy to avoid mutating the original
		newMeta := make(RequestMeta, len(meta)+1)
		for k, v := range meta {
			newMeta[k] = v
		}
		meta = newMeta
	}
	meta[key] = value
	return ContextWithRequestMeta(ctx, meta)
}
