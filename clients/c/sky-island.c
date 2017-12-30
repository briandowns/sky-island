#include <curl/curl.h>
#include <json-c/json.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* client_opt_t holds the configuration for how
 * the client should behave
 */
typedef struct {
  char* endpoint;
  bool skip_host_verify;
  bool skip_peer_verify;
} client_opt_t;

/* response_t contains the response from the Sky Island API */
typedef struct {
  unsigned long timestamp;
  char* data;
} response_t;

/* free_response_t frees the memory used by a response*/
void free_response_t(response_t* res) {
  free(res->data);
  free(res);
}

struct curl_fetch_st {
  char* payload;
  size_t size;
};

/* curl_callback is a callback for curl fetch */
size_t curl_callback(void* contents, size_t size, size_t nmemb, void* userp) {
  size_t realsize = size * nmemb;
  struct curl_fetch_st* p = (struct curl_fetch_st*)userp;

  p->payload = (char*)realloc(p->payload, p->size + realsize + 1);

  if (p->payload == NULL) {
    fprintf(stderr, "ERROR: Failed to expand buffer in curl_callback");
    free(p->payload);
    return -1;
  }

  // copy contents to buffer
  memcpy(&(p->payload[p->size]), contents, realsize);

  // set new buffer size
  p->size += realsize;
  p->payload[p->size] = 0;

  return realsize;
}

/* curl_fetch_url fetches and return url body via curl */
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
  curl_easy_setopt(ch, CURLOPT_USERAGENT, "sky-island/1.0");
  curl_easy_setopt(ch, CURLOPT_TIMEOUT, 5);
  curl_easy_setopt(ch, CURLOPT_FOLLOWLOCATION, 0);

  ret_code = curl_easy_perform(ch);
  return ret_code;
}

/* function is used to make the call to the API and returns back either NULL or
 * a pointer to a response_t. This memory will need to be freed by the caller
 */
response_t* function(client_opt_t* client, const char* url, const char* call) {
  CURL* ch;
  CURLcode ret_code;

  json_object* json;
  enum json_tokener_error jerr = json_tokener_success;

  struct curl_fetch_st curl_fetch;
  struct curl_fetch_st* cf = &curl_fetch;
  struct curl_slist* headers = NULL;

  if ((ch = curl_easy_init()) == NULL) {
    fprintf(stderr, "ERROR: Failed to create curl handle in fetch_session");
    return NULL;
  }

  headers = curl_slist_append(headers, "Accept: application/json");
  headers = curl_slist_append(headers, "Content-Type: application/json");

  json = json_object_new_object();

  json_object_object_add(json, "url", json_object_new_string(url));
  json_object_object_add(json, "call", json_object_new_string(call));

  curl_easy_setopt(ch, CURLOPT_CUSTOMREQUEST, "POST");
  curl_easy_setopt(ch, CURLOPT_HTTPHEADER, headers);
  curl_easy_setopt(ch, CURLOPT_POSTFIELDS, json_object_to_json_string(json));

  if (client->skip_peer_verify) {
    curl_easy_setopt(ch, CURLOPT_SSL_VERIFYPEER, 0L);
  }
  if (client->skip_host_verify) {
    curl_easy_setopt(ch, CURLOPT_SSL_VERIFYHOST, 0L);
  }

  ret_code = curl_fetch_url(ch, client->endpoint, cf);
  curl_easy_cleanup(ch);
  curl_slist_free_all(headers);

  json_object_put(json);

  if (ret_code != CURLE_OK || cf->size < 1) {
    fprintf(stderr, "ERROR: Failed to fetch url (%s) - curl said: %s", url,
            curl_easy_strerror(ret_code));
    return NULL;
  }

  if (cf->payload == NULL) {
    fprintf(stderr, "ERROR: Failed to populate payload");
    free(cf->payload);
    return NULL;
  }

  json = json_tokener_parse_verbose(cf->payload, &jerr);
  free(cf->payload);

  if (jerr != json_tokener_success) {
    fprintf(stderr, "ERROR: Failed to parse json string");
    json_object_put(json);
    return NULL;
  }

  response_t* res = malloc(sizeof(response_t));
  res->timestamp = (unsigned long)time(NULL);
  res->data = strdup(json_object_to_json_string(json));

  // free the memory used by the JSON call
  json_object_put(json);

  return res;
}
