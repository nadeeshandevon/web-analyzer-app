const API_BASE_URL = 'API_BASE_URL_PLACEHOLDER';
const API_KEY = 'API_KEY_PLACEHOLDER';

const app = {
    initApp: function (e) {
        $("#btnAnalyze").on("click", function () {
            app.analyzeWebsite();
        });

        $("#btnReset").on("click", function () {
            app.resetForm();
        });
    },

    analyzeWebsite: function (e) {
        const url = $('#urlInput').val();
        const analyzeBtn = $('#btnAnalyze');
        const loadingSpinner = $('#loadingSpinner');
        const statusText = $('#statusText');
        const resultsContainer = $('#resultsContainer');
        const errorText = $('#errorText');

        if (url === '') {
            errorText.text('Please enter a URL to analyze.');
            return;
        }

        resultsContainer.hide();
        analyzeBtn.disabled = true;
        loadingSpinner.show();
        loadingSpinner.css("display", "inline-block");
        statusText.text('Starting analysis...');
        errorText.text('');

        $.ajax({
            type: "POST",
            url: `${API_BASE_URL}/analyze`,
            headers: {
                'x-api-key': API_KEY
            },
            data: JSON.stringify({ url }),
            dataType: 'json',
            success: analyzeWebsiteSuccess,
            error: analyzeWebsiteError
        });

        function analyzeWebsiteSuccess(data) {
            if (data.analyze_id === undefined || data.analyze_id === null) {
                statusText.innerText = 'Analysis failed.';
                loadingSpinner.hide();
                analyzeBtn.disabled = false;
                return;
            }
            console.log(data);

            const analyzeId = data.analyze_id;

            statusText.text('Analysis in progress...');
            app.pollResults(analyzeId);
        }

        function analyzeWebsiteError(jqXHR, textStatus) {
            console.log(`Error: ${textStatus} - Status: ${jqXHR.status} - Message: ${jqXHR.responseText}`);
            statusText.text(`Analysis failed.`);
            analyzeBtn.disabled = false;
            loadingSpinner.hide();
        }
    },

    pollResults: function (analyzeId) {
        const statusText = $('#statusText');
        const loadingSpinner = $('#loadingSpinner');
        const analyzeBtn = $('#btnAnalyze');
        const errorText = $('#errorText');

        //Poll results every 3 seconds
        const interval = setInterval(function () {
            $.ajax({
                type: "GET",
                url: `${API_BASE_URL}/${analyzeId}/analyze`,
                headers: {
                    'x-api-key': API_KEY
                },
                dataType: 'json',
                success: pollResultsSuccess,
                error: pollResultsError
            });
        }, 3000);

        function pollResultsSuccess(data) {
            if (data.status === 'failed') {
                clearInterval(interval);
                errorText.text(data.error_description)
                statusText.text('Analysis failed.');
                loadingSpinner.hide();
                analyzeBtn.disabled = false;
                return;
            }
            console.log(data);

            if (data.status === 'success') {
                clearInterval(interval);
                app.renderResults(data);
                statusText.text('Analysis completed!');
                loadingSpinner.hide();
                analyzeBtn.disabled = false;
            } else if (data.status === 'failed') {
                clearInterval(interval);
                statusText.text('Analysis failed.');
                loadingSpinner.hide();
                analyzeBtn.disabled = false;
            }
        }

        function pollResultsError(jqXHR, textStatus) {
            console.log(`Error: ${textStatus} - Status: ${jqXHR.status} - Message: ${jqXHR.responseText}`);
            statusText.text(`Failed to fetch analysis results.`);
            analyzeBtn.disabled = false;
            loadingSpinner.hide();
        }
    },

    renderResults: function (data) {
        const resultsContainer = $('#resultsContainer');
        resultsContainer.show();
        $('#resUrl').text(data.url);
        $('#resHtmlVersion').text(data.html_version);
        $('#resTitle').text(data.title);
        $('#resLoginForm').text(data.login_form);
        $('#resHeadings').empty();
        if (data.headings) {
            for (const [tag, count] of Object.entries(data.headings)) {
                $('#resHeadings').append(`<li><strong>${tag.toUpperCase()}:</strong> ${count}</li>`);
            }
        }
        $('#resInternalLinks').text(data.links.internal);
        $('#resExternalLinks').text(data.links.external);
        $('#resInaccessibleLinks').text(data.links.inaccessible);
        $('#resLoginForm').text(data.has_login_form ? 'Yes' : 'No');
    },

    resetForm: function () {
        $('#urlInput').val('');
        $('#statusText').text('');
        $('#errorText').text('');
        $('#resultsContainer').hide();
        $('#resLoginForm').text('');
        $('#btnAnalyze').prop('disabled', false);
        $('#loadingSpinner').hide();
    }
}