#include <curl/curl.h>
#include <json-c/json.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct client_t {
  char* endpoint;
};

struct response_t {
  unsigned long timestamp;
  void* data;
};

struct curl_fetch_st {
  char* payload;
  size_t size;
};

/* callback for curl fetch */
size_t curl_callback(void* contents, size_t size, size_t nmemb, void* userp) {
  size_t realsize = size * nmemb; /* calculate buffer size */
  struct curl_fetch_st* p =
      (struct curl_fetch_st*)userp; /* cast pointer to fetch struct */

  /* expand buffer */
  p->payload = (char*)realloc(p->payload, p->size + realsize + 1);

  if (p->payload == NULL) {
    fprintf(stderr, "ERROR: Failed to expand buffer in curl_callback");
    free(p->payload);
    return -1;
  }

  /* copy contents to buffer */
  memcpy(&(p->payload[p->size]), contents, realsize);

  /* set new buffer size */
  p->size += realsize;
  p->payload[p->size] = 0;

  return realsize;
}

/* fetch and return url body via curl */
CURLcode curl_fetch_url(CURL* ch,
                        const char* url,
                        struct curl_fetch_st* fetch) {
  CURLcode ret_code;

  fetch->payload = (char*)calloc(1, sizeof(fetch->payload));

  if (fetch->payload == NULL) {
    fprintf(stderr, "ERROR: Failed to allocate payload in curl_fetch_url");
    return CURLE_FAILED_INIT;
  }

  fetch->size = 0;

  curl_easy_setopt(ch, CURLOPT_URL, url);
  curl_easy_setopt(ch, CURLOPT_WRITEFUNCTION, curl_callback);
  curl_easy_setopt(ch, CURLOPT_WRITEDATA, (void*)fetch);
  curl_easy_setopt(ch, CURLOPT_USERAGENT, "sky-island-agent/1.0");
  curl_easy_setopt(ch, CURLOPT_TIMEOUT, 5);
  curl_easy_setopt(ch, CURLOPT_FOLLOWLOCATION, 0);

  ret_code = curl_easy_perform(ch);

  return ret_code;
}

static int function(struct client_t* client,
                    const char* url,
                    const char* call) {
  CURL* ch;
  CURLcode ret_code;

  json_object* json;
  enum json_tokener_error jerr = json_tokener_success;

  struct curl_fetch_st curl_fetch;
  struct curl_fetch_st* cf = &curl_fetch;
  struct curl_slist* headers = NULL;

  if ((ch = curl_easy_init()) == NULL) {
    fprintf(stderr, "ERROR: Failed to create curl handle in fetch_session");
    return -1;
  }

  headers = curl_slist_append(headers, "Accept: application/json");
  headers = curl_slist_append(headers, "Content-Type: application/json");

  json = json_object_new_object();

  json_object_object_add(json, "url", json_object_new_string(url));
  json_object_object_add(json, "call", json_object_new_string(call));

  curl_easy_setopt(ch, CURLOPT_CUSTOMREQUEST, "POST");
  curl_easy_setopt(ch, CURLOPT_HTTPHEADER, headers);
  curl_easy_setopt(ch, CURLOPT_POSTFIELDS, json_object_to_json_string(json));

  ret_code = curl_fetch_url(ch, client->endpoint, cf);
  curl_easy_cleanup(ch);
  curl_slist_free_all(headers);

  json_object_put(json);

  if (ret_code != CURLE_OK || cf->size < 1) {
    fprintf(stderr, "ERROR: Failed to fetch url (%s) - curl said: %s", url,
            curl_easy_strerror(ret_code));
    return -1;
  }

  if (cf->payload == NULL) {
    fprintf(stderr, "ERROR: Failed to populate payload");
    free(cf->payload);
    return -1;
  }

  json = json_tokener_parse_verbose(cf->payload, &jerr);
  free(cf->payload);

  if (jerr != json_tokener_success) {
    fprintf(stderr, "ERROR: Failed to parse json string");
    json_object_put(json);
    return -1;
  }

  printf("%s\n", json_object_to_json_string(json));

  json_object_put(json);

  return 0;
}