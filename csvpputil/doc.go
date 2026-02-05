// Package csvpputil provides utility functions for converting CSV++ data
// to other formats such as JSON and YAML.
//
// # JSON Streaming Output
//
// For large JSON files, use JSONArrayWriter to stream records directly to the output:
//
//	w := csvpputil.NewJSONArrayWriter(out, headers)
//	for {
//	    record, err := reader.Read()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        return err
//	    }
//	    if err := w.Write(record); err != nil {
//	        return err
//	    }
//	}
//	if err := w.Close(); err != nil {
//	    return err
//	}
//
// # YAML Streaming Output
//
// For YAML output, use YAMLArrayWriter:
//
//	w := csvpputil.NewYAMLArrayWriter(out, headers)
//	for _, record := range records {
//	    if err := w.Write(record); err != nil {
//	        return err
//	    }
//	}
//	if err := w.Close(); err != nil {
//	    return err
//	}
//
// # Convenience Functions
//
// For small to medium datasets, use the Marshal or Write functions:
//
//	data, err := csvpputil.MarshalJSON(headers, records)
//	data, err := csvpputil.MarshalYAML(headers, records)
//
//	err := csvpputil.WriteJSON(w, headers, records)
//	err := csvpputil.WriteYAML(w, headers, records)
package csvpputil
