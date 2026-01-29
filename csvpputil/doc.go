// Package csvpputil provides utility functions for converting CSV++ data
// to other formats such as JSON and YAML.
//
// # JSON Streaming Encoding
//
// For large JSON files, use JSONEncoder to stream records directly to the output:
//
//	enc := csvpputil.NewJSONEncoder(w, headers)
//	for {
//	    record, err := reader.Read()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        return err
//	    }
//	    if err := enc.Encode(record); err != nil {
//	        return err
//	    }
//	}
//	if err := enc.Close(); err != nil {
//	    return err
//	}
//
// # Convenience Functions
//
// For small to medium datasets, use the Marshal functions:
//
//	data, err := csvpputil.MarshalJSON(headers, records)
//	data, err := csvpputil.MarshalYAML(headers, records)
package csvpputil
