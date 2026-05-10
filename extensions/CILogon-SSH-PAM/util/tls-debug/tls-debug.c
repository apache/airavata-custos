#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>		/* strerror */
#include <curl/curl.h>


/** Max number of bytes to fetch from server */

#define MAX_DATA 1200


/** setup curl
 * \param uninitialised curl ptr
 * \param trust anchor path
 * \return errno or 0 if successful
 */

int setup(CURL **, char const *ta_path);



struct call_data {
    char errbuf[CURL_ERROR_SIZE];
    char data[MAX_DATA];
    size_t fill;
};


/** call a URL
 * \param initialised curl structure
 * \param url to connect to
 * \param uninitialised call_data structure will be filled in
 * \return curl error code
 */

CURLcode call(CURL *curl, char const *url, struct call_data *);


int
main(int argc, char **argv)
{
    CURL *curl = 0;
    struct call_data data;
    // Check if we have a trust anchor
    if(argc != 3) {
	fprintf(stderr, "Usage: %s <trust-anchor> <remote_http_url>\n", argv[0]);
	exit(1);
    }
    // Set up curl and configure the trust anchor
    int ret = setup(&curl, argv[1]);
    if(ret != 0) {
	fprintf(stderr, "Setup error: %s\n", strerror(ret));
	exit(1);
    }
    if((ret = call(curl, argv[2], &data)) != 0) {
	fprintf(stderr, "Failed call to %s: %s\n%s\n",
		argv[2],
		curl_easy_strerror(ret),
		data.errbuf);
	return 1;
    }
    return 0;
}




size_t
callback(char const *contents, size_t size, size_t nmemb, void *user)
{
    struct call_data *data = (struct call_data *)user;
    size_t to_read = size*nmemb;
    if(to_read > MAX_DATA - data->fill)
	to_read = MAX_DATA - data->fill;
    if(to_read > 0) {
	memcpy(&(data->data[data->fill]), contents, to_read);
	data->fill += to_read;
    }
    return to_read;
}


int
setup(CURL **curl, char const *ta_path)
{
    struct stat fs;
    if(stat(ta_path, &fs))
	return errno;
    *curl = curl_easy_init();
    if(!*curl) {
	fprintf(stderr, "Cannot init curl\n");
	return EINVAL;
    }
    
    if(curl_easy_setopt(*curl, CURLOPT_SSL_VERIFYPEER, 1L) != CURLE_OK) {
	fprintf(stderr, "Failed to set verify peer flag\n");
	return EINVAL;
    }

    if(curl_easy_setopt(*curl, CURLOPT_SSL_VERIFYHOST, 2L) != CURLE_OK) {
	fprintf(stderr, "Failed to set verify peer flag\n");
	return EINVAL;
    }


    switch(fs.st_mode & S_IFMT) {
    case S_IFREG:
	if(curl_easy_setopt(*curl, CURLOPT_CAINFO, ta_path) != CURLE_OK) {
	    fprintf(stderr, "Failed to set CA **bundle** to %s\n", ta_path);
	    return EINVAL;
	}
	break;
    case S_IFDIR:
	if(curl_easy_setopt(*curl, CURLOPT_CAPATH, ta_path) != CURLE_OK) {
	    fprintf(stderr, "Failed to set CA **path** to %s\n", ta_path);
	    return EINVAL;
	}
	break;
    default:
	fprintf(stderr, "Unknown or unsupported data type %s: %d\n", ta_path, fs.st_mode & S_IFMT);
	return EINVAL;
    }

    return 0;
}



CURLcode
call(CURL *curl, char const *url, struct call_data *data)
{
    data->fill = 0;
    CURLcode ret;
    if((ret = curl_easy_setopt(curl, CURLOPT_URL, url)) != CURLE_OK) {
	fputs("Failed to set URL\n", stderr);
	return ret;
    }
    if((ret = curl_easy_setopt(curl, CURLOPT_VERBOSE, 1)) != CURLE_OK) {
	fputs("Warning: failed to set verbose\n", stderr);
    }
    if((ret = curl_easy_setopt(curl, CURLOPT_HEADER, 1)) != CURLE_OK) {
	fputs("Warning: failed to set header out\n", stderr);
    }
    if((ret = curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, callback)) != CURLE_OK) {
	fputs("Failed to set callback function\n", stderr);
	return ret;
    }
    if((ret = curl_easy_setopt(curl, CURLOPT_WRITEDATA, data->data)) != CURLE_OK) {
	fputs("Failed to set callback buffer\n", stderr);
	return ret;
    }

    /* The Moment of Truth... */
    return curl_easy_perform(curl);
}
